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
	chooseOne       = "Choose one"
	importList      = "Import board game list"
	randomCompare   = "Random compare"
	top20           = "Top 20"
	settings        = "Settings"
	cancel          = "Cancel"
	twoStepCompare  = "Two step compare"
	another         = "Another"
	defaultCategory = "Default Category"
	selectCategory  = "Select Active Category"
)

var defaultLists = map[string]gobgg.CollectionType{
	"Wishlist": gobgg.CollectionTypeWishList,
	"Own":      gobgg.CollectionTypeOwn,
	"Played":   gobgg.CollectionTypePlayed,
	"Rated":    gobgg.CollectionTypeRated,
}

type singleUser struct {
	userID  int64
	ranker  ranking.Ranker
	storage db.Storage

	left  *db.Item
	right *db.Item

	category *db.Category

	config struct {
		twoStepCompare bool
	}

	state stateFunc
}

func (su *singleUser) resetText(_ context.Context, text string) telegram.Response {
	su.left, su.right = nil, nil
	su.state = su.startState
	return telegram.NewButtonResponse(text,
		importList,
		randomCompare,
		top20,
		selectCategory,
		settings,
	)
}

func (su *singleUser) Reset(ctx context.Context) telegram.Response {
	var err error
	su.category, err = su.storage.GetCategoryByName(ctx, su.userID, "Wishlist")
	if err != nil {
		log.Print(err)
	}
	return su.resetText(ctx, chooseOne)
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
	case importList:
		return telegram.NewTextResponse("Your user name:", true), su.importState
	case top20:
		items, err := su.page(context.Background(), 1, 20)
		if err != nil {
			return su.errState(ctx, err)
		}

		text := "Your top ten list:\n"
		for i := range items {
			text += fmt.Sprintf("%d => %s\n%s\n", items[i].Rank, items[i].Name, items[i].URL)
		}

		return su.resetText(ctx, text), su.startState
	case randomCompare:
		if err := su.getComparableItems(context.Background()); err != nil {
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

		buttons = append(buttons, cancel)
		items3 := telegram.NewButtonResponse("Choose one option or enter a number between 0-100:", buttons...)
		return telegram.NewMultiResponse(item1, item2, items3), su.stateBattle
	case selectCategory:
		return su.setCategory(ctx, "")
	case settings:
		return su.setConfig(ctx, "")
	default:
		return telegram.NewTextResponse("invalid input", false), su.startState
	}
}

func (su *singleUser) setCategory(ctx context.Context, message string) (telegram.Response, stateFunc) {
	cats, err := su.storage.Categories(ctx, su.userID)
	if err != nil {
		return su.errState(ctx, err)
	}

	if len(cats) == 0 {
		return su.resetText(ctx, "No category, import board games first"), su.startState
	}

	var (
		data = make([]string, 0, len(cats))
	)

	if message == "" {
		for i := range cats {
			data = append(data, cats[i].Name)
		}

		return telegram.NewButtonResponse("Select the category", data...), su.setCategory
	}

	for i := range cats {
		if message == cats[i].Name {
			su.category = cats[i]
			return su.resetText(ctx, fmt.Sprintf("Active category is %s", su.category.Name)), su.startState
		}
	}

	return su.resetText(ctx, "Invalid category name"), su.startState
}

func (su *singleUser) setConfig(ctx context.Context, message string) (telegram.Response, stateFunc) {
	if message == "" {
		status := "ON"
		if su.config.twoStepCompare {
			status = "OFF"
		}
		return telegram.NewButtonResponse("Set config", fmt.Sprintf("%s %s", twoStepCompare, status), cancel), su.setConfig
	}
	switch message {
	case cancel:
		return su.resetText(ctx, "Nothing was changed"), su.startState
	case fmt.Sprintf("%s ON", twoStepCompare):
		su.config.twoStepCompare = true
	case fmt.Sprintf("%s OFF", twoStepCompare):
		su.config.twoStepCompare = false
	default:
		return su.errState(ctx, errors.New("invalid configuration"))
	}

	return su.resetText(ctx, "Config set was successful"), su.startState
}

func (su *singleUser) fallbackToFloat(ctx context.Context, message string) (telegram.Response, stateFunc) {
	score, err := strconv.ParseFloat(message, 64)
	if err != nil {
		return telegram.NewTextResponse("Input one option", false), su.stateBattle
	}
	if score < 0 || score > 100 {
		return telegram.NewTextResponse("Score should be between [0-100] or input one option", false), su.stateBattle
	}

	return su.rankMessage(ctx, score/100)
}

func (su *singleUser) afterBattle(ctx context.Context, message string) (telegram.Response, stateFunc) {
	switch message {
	case another:
		return su.startState(ctx, randomCompare)
	default:
		return su.Reset(ctx), su.startState
	}
}

func (su *singleUser) rankMessage(ctx context.Context, score float64) (telegram.Response, stateFunc) {
	text, err := su.setRank(ctx, score)
	if err != nil {
		return su.errState(ctx, err)
	}

	return telegram.NewButtonResponse(text, another, cancel), su.afterBattle

}

func (su *singleUser) stateBattle(ctx context.Context, message string) (telegram.Response, stateFunc) {
	if message == cancel {
		return su.Reset(ctx), su.startState
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

	if !su.config.twoStepCompare {
		return su.startState(ctx, randomCompare)
	}
	return su.rankMessage(ctx, score)
}

func (su *singleUser) getButtonText() (map[string]float64, []string, error) {
	if su.left == nil || su.right == nil {
		return nil, nil, errors.New("nothing selected")
	}
	items := []string{
		fmt.Sprintf("%s is 100%% winner", su.left.Name),
		fmt.Sprintf("%s is 75%% winner", su.left.Name),
		"Equal",
		fmt.Sprintf("%s is 75%% winner", su.right.Name),
		fmt.Sprintf("%s is 100%% winner", su.right.Name),
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
		su.Reset(ctx)), su.startState
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
			fmt.Sprintf("%d items was in your %q list, %d was new", count, cat.Name, count-old))
	}

	return strings.Join(result, "\n"), nil
}

// NewChat creates a new telegram chat
func NewChat(userID int64, ranker ranking.Ranker, storage db.Storage) *singleUser {
	return &singleUser{
		userID:  userID,
		ranker:  ranker,
		storage: storage,
	}
}
