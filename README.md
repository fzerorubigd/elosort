# Wishlist battle

**Note: this is a draft code. I just built this for my usage, so no test, 
general-purpose, or great program (it is a shitty code, I know)! it works only for board games, but you 
can add other things easily**

## Why?

I can not choose which board game to buy. That's my curse, I guess. I spent hours trying to decide what to buy in my favorite shop (With my limited budget, of course).

So I built this; this is a ranking system.
You can import your things (I implement board games only), and then 
you can compare them whenever you have time.  

The ranking is based on Elo ranking for chess. So it shows you a menu,
with five buttons, you decide which one the winner is (also you can 
mark them as 75% percent or any arbitrary percentage winner), 
then it updates the rank for both based on the old rank and new score.

It might ask you again later on the same item, but that's ok. I go for my feeling toward that items right now, which one I prefer to play now.

The result was great so far! :)  

## How to build

This project uses Magefile. run `go run ./main.go build` (or its shortcut `make build`)

Use telegram botfother to create a new bot and pass the token with `-token` switch or `TELEGRAM_BOT_TOKEN` environment variable. 
You can also set the database path, it will create one if it is empty.  

