.DEFAULT_GOAL = help
%:
	@go run main.go $@

help:
	@go run main.go