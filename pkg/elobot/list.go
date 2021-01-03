package elobot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fzerorubigd/gobgg"
	"github.com/pkg/errors"

	"github.com/fzerorubigd/elosort/pkg/db"
	"github.com/fzerorubigd/elosort/pkg/ranking"
	"github.com/fzerorubigd/elosort/pkg/telegram"
)

type stateFunc func(ctx context.Context, message string) (telegram.Response, stateFunc)

const (
	chooseOne            = "Select one option:"
	importList           = "Import board game list from bgg"
	randomCompare        = "Random compare"
	manageItems          = "Manage items"
	top20                = "Top 20"
	settings             = "Settings"
	cancel               = "Cancel"
	twoStepCompare       = "Two step compare"
	setLanguage          = "Set language"
	another              = "Next"
	selectCategory       = "Select Active Category"
	yesAction            = "Yes"
	noAction             = "No"
	yourUserName         = "Your BGG username:"
	yourTopTenList       = "Your top ten list (Category: %s):\n"
	chooseOption         = "This is your %q \nChoose one option or enter a number between 0-100:"
	invalidInput         = "Invalid input"
	importFirst          = "No category, import board games first"
	nothingWasChanged    = "Nothing was changed"
	selectActiveCategory = "Select the category, current active is: "
	configSaved          = "Config saved"
	wishList             = "Wishlist"
	own                  = "Own"
	played               = "Played"
	rated                = "Rated"
	unknown              = "Unknown"
	deleteItem           = "Delete %q"
	compareString        = "%s is %d%% winner"
	equal                = "Equal"
	itemsInYourList      = "%d items was in your %q list, %d was new"
	areYouSure           = "Are you sure? this can't be undone"
)

var defaultLists = map[string]gobgg.CollectionType{
	wishList: gobgg.CollectionTypeWishList,
	own:      gobgg.CollectionTypeOwn,
	played:   gobgg.CollectionTypePlayed,
	rated:    gobgg.CollectionTypeRated,
}

type singleUser struct {
	userID  int64
	ranker  ranking.Ranker
	storage db.Storage

	left  *db.Item
	right *db.Item

	category *db.Category

	config db.UserConfig

	state stateFunc
}

func (su *singleUser) translate(in string) string {
	if su.config.Language == "" {
		su.config.Language = "En"
	}

	return t(in, su.config.Language)
}

func (su *singleUser) translateArray(in ...string) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = su.translate(in[i])
	}

	return out
}

func (su *singleUser) resetText(_ context.Context, text string) telegram.Response {
	su.left, su.right = nil, nil
	su.state = su.startState
	return telegram.NewButtonResponse(text,
		su.translateArray(importList,
			randomCompare,
			top20,
			selectCategory,
			settings)...,
	)
}

func (su *singleUser) Reset(ctx context.Context) telegram.Response {
	var err error
	if su.config.DefaultCatID == 0 {
		su.category, err = su.storage.GetCategoryByName(ctx, su.userID, wishList)
		if err != nil {
			log.Print(err)
		}

		su.config.DefaultCatID = su.category.GetID()
		if err := su.storage.UpdateConfig(ctx, su.userID, &su.config); err != nil {
			log.Print(err)
		}
	} else {
		su.category, err = su.storage.GetCategoryByID(ctx, su.config.DefaultCatID)
		if err != nil {
			log.Print(err)
		}
	}

	return su.resetText(ctx, su.translate(chooseOne))
}

func (su *singleUser) Process(ctx context.Context, message string) telegram.Response {
	if su.state == nil {
		return su.Reset(ctx)
	}

	resp, state := su.state(ctx, message)
	su.state = state

	return resp
}

func (su *singleUser) errState(ctx context.Context, err error) (telegram.Response, stateFunc) {
	return su.resetText(ctx, fmt.Sprintf("Error: %q", err.Error())), su.startState
}

func (su *singleUser) startState(ctx context.Context, message string) (telegram.Response, stateFunc) {
	switch message {
	case su.translate(importList):
		return telegram.NewTextResponse(su.translate(yourUserName), true), su.importState
	case su.translate(top20):
		items, err := su.page(context.Background(), 1, 20)
		if err != nil {
			return su.errState(ctx, err)
		}
		cat := su.translate(unknown)
		if su.category != nil {
			cat = su.translate(su.category.Name)
		}
		text := fmt.Sprintf(su.translate(yourTopTenList), cat)
		for i := range items {
			text += fmt.Sprintf("%d => %s\n%s\n", items[i].Rank, items[i].Name, items[i].URL)
		}

		return su.resetText(ctx, text), su.startState
	case su.translate(randomCompare):
		if err := su.getComparableItems(ctx); err != nil {
			return su.errState(ctx, err)
		}

		item1 := telegram.NewTextResponse(
			fmt.Sprintf("%d => %s \n%s", su.left.Rank, su.left.Name, su.left.URL),
			true,
		)

		item2 := telegram.NewTextResponse(
			fmt.Sprintf("%d => %s \n%s", su.right.Rank, su.right.Name, su.right.URL),
			true,
		)

		_, buttons, err := su.getButtonText()
		if err != nil {
			return su.errState(ctx, err)
		}

		buttons = append(buttons, su.translateArray(manageItems, cancel)...)
		cat := su.translate(unknown)
		if su.category != nil {
			cat = su.translate(su.category.Name)
		}
		items3 := telegram.NewButtonResponse(
			fmt.Sprintf(su.translate(chooseOption), cat), buttons...)
		return telegram.NewMultiResponse(item1, item2, items3), su.stateBattle
	case su.translate(selectCategory):
		return su.setCategory(ctx, "")
	case su.translate(settings):
		return su.setConfig(ctx, "")
	default:
		return telegram.NewTextResponse(su.translate(invalidInput), false), su.startState
	}
}

func (su *singleUser) setCategory(ctx context.Context, message string) (telegram.Response, stateFunc) {
	cats, err := su.storage.Categories(ctx, su.userID)
	if err != nil {
		return su.errState(ctx, err)
	}

	if len(cats) == 0 {
		return su.resetText(ctx, su.translate(importFirst)), su.startState
	}

	var (
		data = make([]string, 0, len(cats)+1)
	)

	if message == "" {
		for i := range cats {
			data = append(data, cats[i].Name)
		}
		cat := su.translate(unknown)
		if su.category != nil {
			cat = su.translate(su.category.Name)
		}
		data = append(data, su.translate(cancel))
		return telegram.NewButtonResponse(su.translate(selectActiveCategory)+cat, data...), su.setCategory
	}

	if message == su.translate(cancel) {
		return su.resetText(ctx, su.translate(nothingWasChanged)), su.startState
	}

	for i := range cats {
		if message == su.translate(cats[i].Name) {
			su.category = cats[i]
			su.config.DefaultCatID = cats[i].ID
			if err := su.storage.UpdateConfig(ctx, su.userID, &su.config); err != nil {
				log.Print(err)
			}
			return su.resetText(ctx, fmt.Sprintf(su.translate("Active category is %s"), su.category.Name)), su.startState
		}
	}

	return su.resetText(ctx, su.translate("Invalid category name")), su.startState
}

func (su *singleUser) setConfig(ctx context.Context, message string) (telegram.Response, stateFunc) {
	if message == "" {
		status := "ON"
		if su.config.ShowTwoStep {
			status = "OFF"
		}
		lang := "Fa"
		if su.config.Language == "Fa" {
			lang = "En"
		}
		return telegram.NewButtonResponse(
				su.translate("Set config"),
				fmt.Sprintf("%s %s", su.translate(twoStepCompare), status),
				fmt.Sprintf("%s %s", su.translate(setLanguage), lang),
				su.translate(cancel)),
			su.setConfig
	}
	switch message {
	case su.translate(cancel):
		return su.resetText(ctx, su.translate(nothingWasChanged)), su.startState
	case fmt.Sprintf("%s ON", su.translate(twoStepCompare)):
		su.config.ShowTwoStep = true
	case fmt.Sprintf("%s OFF", su.translate(twoStepCompare)):
		su.config.ShowTwoStep = false
	case fmt.Sprintf("%s Fa", su.translate(setLanguage)):
		su.config.Language = "Fa"
	case fmt.Sprintf("%s En", su.translate(setLanguage)):
		su.config.Language = "En"
	default:
		return su.errState(ctx, errors.New("invalid configuration"))
	}

	if err := su.storage.UpdateConfig(ctx, su.userID, &su.config); err != nil {
		log.Print(err)
	}

	return su.resetText(ctx, su.translate(configSaved)), su.startState
}

func (su *singleUser) fallbackToFloat(ctx context.Context, message string) (telegram.Response, stateFunc) {
	score, err := strconv.ParseFloat(message, 64)
	if err != nil {
		return telegram.NewTextResponse(su.translate(chooseOne), false), su.stateBattle
	}
	if score < 0 || score > 100 {
		return telegram.NewTextResponse(chooseOption, false), su.stateBattle
	}

	return su.rankMessage(ctx, score/100)
}

func (su *singleUser) afterBattle(ctx context.Context, message string) (telegram.Response, stateFunc) {
	switch message {
	case su.translate(another):
		return su.startState(ctx, su.translate(randomCompare))
	default:
		return su.resetText(ctx, su.translate(chooseOne)), su.startState
	}
}

func (su *singleUser) rankMessage(ctx context.Context, score float64) (telegram.Response, stateFunc) {
	text, err := su.setRank(ctx, score)
	if err != nil {
		return su.errState(ctx, err)
	}
	if !su.config.ShowTwoStep {
		return su.afterBattle(ctx, su.translate(another))
	}
	return telegram.NewButtonResponse(text, su.translateArray(another, cancel)...), su.afterBattle
}

func (su *singleUser) manageState(ctx context.Context, message string) (telegram.Response, stateFunc) {
	buttons := []string{
		fmt.Sprintf(su.translate(deleteItem), su.left.Name),
		fmt.Sprintf(su.translate(deleteItem), su.right.Name),
		su.translate(cancel),
	}
	if message == "" {
		return telegram.NewButtonResponse("Select one to remove: ", buttons...), su.manageState
	}

	if message == su.translate(cancel) {
		return su.startState(ctx, su.translate(randomCompare))
	}

	var active *db.Item
	if message == buttons[0] {
		active = su.left
	} else if message == buttons[1] {
		active = su.right
	}

	if active == nil {
		return su.startState(ctx, su.translate(randomCompare))
	}

	return telegram.NewButtonResponse(areYouSure, su.translateArray(yesAction, noAction)...),
		func(ctx context.Context, message string) (telegram.Response, stateFunc) {
			if message == su.translate(yesAction) {
				if err := su.storage.Remove(ctx, active.ID); err != nil {
					log.Print(err)
				}
			}
			return su.startState(ctx, su.translate(randomCompare))
		}
}

func (su *singleUser) stateBattle(ctx context.Context, message string) (telegram.Response, stateFunc) {
	if message == su.translate(cancel) {
		return su.resetText(ctx, su.translate(chooseOne)), su.startState
	}

	if message == su.translate(manageItems) {
		return su.manageState(ctx, "")
	}

	texts, _, err := su.getButtonText()
	if err != nil {
		return su.errState(ctx, err)
	}

	score, ok := texts[message]
	if !ok {
		// fallback to numbers
		return su.fallbackToFloat(ctx, message)
	}

	return su.rankMessage(ctx, score)
}

func (su *singleUser) getButtonText() (map[string]float64, []string, error) {
	if su.left == nil || su.right == nil {
		return nil, nil, errors.New("nothing selected")
	}
	items := []string{
		fmt.Sprintf(su.translate(compareString), su.left.Name, 100),
		fmt.Sprintf(su.translate(compareString), su.left.Name, 75),
		su.translate(equal),
		fmt.Sprintf(su.translate(compareString), su.right.Name, 100),
		fmt.Sprintf(su.translate(compareString), su.right.Name, 75),
	}

	m := make(map[string]float64)
	for i := range items {
		idx := 1 - float64(i)*0.25
		m[items[i]] = idx
	}
	return m, items, nil
}

func (su *singleUser) importState(ctx context.Context, message string) (telegram.Response, stateFunc) {
	str, err := su.importList(ctx, message)
	if err != nil {
		return su.errState(ctx, err)
	}

	return telegram.NewMultiResponse(telegram.NewTextResponse(str, true),
		su.resetText(ctx, su.translate(chooseOne))), su.startState
}

func (su *singleUser) getComparableItems(ctx context.Context) error {
	items, err := su.storage.Random(ctx, su.userID, su.category.GetID(), 2)
	if err != nil {
		return errors.Wrap(err, "failed to get random items")
	}

	if len(items) != 2 {
		return errors.New("there are not enough item in db")
	}

	su.left = items[0]
	su.right = items[1]

	return nil
}

func (su *singleUser) setRank(ctx context.Context, leftScore float64) (string, error) {
	if su.left == nil || su.right == nil {
		return "", errors.New("no active comparison")
	}
	if leftScore < 0 || leftScore > 1 {
		return "", errors.New("invalid score")
	}

	left, right, err := su.ranker.Calculate(su.left.Rank, su.right.Rank, leftScore, 1-leftScore)
	if err != nil {
		return "", errors.Wrap(err, "ranker failed")
	}

	text := fmt.Sprintf("%s (%d)=> %d to %d\n", su.left.Name, int(leftScore*100), su.left.Rank, left)
	if err := su.storage.SetRank(ctx, su.left.ID, left); err != nil {
		return "", errors.Wrap(err, "update rank failed")
	}
	text += fmt.Sprintf("%s (%d)=> %d to %d\n", su.right.Name, int((1-leftScore)*100), su.right.Rank, right)
	if err := su.storage.SetRank(ctx, su.right.ID, right); err != nil {
		return "", errors.Wrap(err, "update rank failed")
	}

	su.left, su.right = nil, nil
	return text, nil
}

func (su *singleUser) page(ctx context.Context, page, count int) ([]*db.Item, error) {
	return su.storage.Items(ctx, su.userID, su.category.GetID(), page, count)
}

func (su *singleUser) findList(ctx context.Context, all []*db.Category, list string) (*db.Category, error) {
	for i := range all {
		if all[i].Name == list {
			return all[i], nil
		}
	}

	cat := db.Category{
		UserID:      su.userID,
		Name:        list,
		Description: "",
	}

	return su.storage.CreateCategory(ctx, &cat)
}

func (su *singleUser) importList(ctx context.Context, username string) (string, error) {
	bgg := gobgg.NewBGGClient()

	var result []string
	cats, err := su.storage.Categories(ctx, su.userID)
	if err != nil {
		return "", err
	}

	for i := range defaultLists {
		things, err := bgg.GetCollection(ctx, username, gobgg.SetCollectionTypes(defaultLists[i]))
		if err != nil {
			return "", err
		}

		if len(things) < 0 {
			continue
		}

		cat, err := su.findList(ctx, cats, i)
		if err != nil {
			return "", err
		}

		count, old := len(things), 0
		for i := range things {
			item := db.Item{
				UserID:      su.userID,
				Category:    cat.ID,
				Name:        things[i].Name,
				Description: strings.Trim(things[i].Description, "\n\t "),
				URL:         fmt.Sprintf("https://boardgamegeek.com/boardgame/%d/", things[i].ID),
				Image:       "",
			}

			if _, err := su.storage.Create(ctx, &item); err != nil {
				log.Println("Already there")
				old++
			}
		}

		result = append(result,
			fmt.Sprintf(su.translate(itemsInYourList), count, su.translate(cat.Name), count-old))
	}

	return strings.Join(result, "\n"), nil
}

// NewChat creates a new telegram chat
func NewChat(ctx context.Context, userID int64, ranker ranking.Ranker, storage db.Storage) *singleUser {
	usr, err := storage.UserByID(ctx, userID)
	if err != nil {
		log.Print("New user?")
		usr = &db.User{
			ID:     userID,
			Config: &db.UserConfig{},
		}
		if err := storage.CreateUser(ctx, usr); err != nil {
			log.Print(err)
		}
	}
	return &singleUser{
		userID:  userID,
		ranker:  ranker,
		storage: storage,
		config:  *usr.Config,
	}
}
