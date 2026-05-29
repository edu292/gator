package commands

import (
	"context"
	"errors"
	"fmt"

	"gator/internal/config"
	"gator/internal/database"
	"gator/internal/rss"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var ErrFatal = errors.New("fatal")

type command struct {
	name string
	args []string
}

func NewCommand(args []string) (command, error) {
	if lenArgs := len(args); lenArgs < 2 {
		return command{}, fmt.Errorf("%w: Too few arguments. Expected at least 2. Got: %d", ErrFatal, lenArgs)
	}

	return command{name: args[1], args: args[2:]}, nil
}

func parseDBErr(err error) error {
	if err == nil {
		return nil
	}
	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		switch pqErr.Code {
		case "23505":
			return fmt.Errorf("duplicate entry: %s", pqErr.Detail)
		case "23503":
			return fmt.Errorf("missing reference: %s", pqErr.Detail)
		default:
			return fmt.Errorf("db error %s: %s", pqErr.Code, pqErr.Message)
		}
	}
	return err
}

func handlerLogin(ctx context.Context, s *config.State, cmd command) error {
	userName := cmd.args[0]
	user, err := s.DB.GetUserByName(ctx, userName)
	if err != nil {
		return fmt.Errorf("%w: get user failed: %w", ErrFatal, parseDBErr(err))
	}

	s.Cfg.SetUser(user.ID)
	fmt.Printf("User set: %s\n", userName)
	return nil
}

func handlerRegister(ctx context.Context, s *config.State, cmd command) error {
	userName := cmd.args[0]
	userID := uuid.New()
	_, err := s.DB.CreateUser(ctx, database.CreateUserParams{
		Name: userName,
		ID:   userID,
	})
	if err != nil {
		return fmt.Errorf("%w: register failed: %w", ErrFatal, parseDBErr(err))
	}

	s.Cfg.SetUser(userID)
	fmt.Printf("User created: %s\n", userName)
	return nil
}

func handlerReset(ctx context.Context, s *config.State, cmd command) error {
	if err := s.DB.ResetUsers(ctx); err != nil {
		return fmt.Errorf("%w: reset failed: %w", ErrFatal, parseDBErr(err))
	}

	fmt.Println("Reset successful")
	return nil
}

func handlerUsers(ctx context.Context, s *config.State, cmd command) error {
	users, err := s.DB.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("%w: get users failed: %w", ErrFatal, parseDBErr(err))
	}

	for _, user := range users {
		fmt.Print("* ", user.Name)
		if user.ID == s.Cfg.CurrentUserID {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	return nil
}

func handlerAgg(ctx context.Context, s *config.State, cmd command) error {
	feed, err := rss.FetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("%w: fetch feed failed: %w", ErrFatal, err)
	}
	fmt.Println(*feed)
	return nil
}

func handlerAddFeed(ctx context.Context, s *config.State, cmd command, user database.User) error {
	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.DB.CreateFeed(ctx, database.CreateFeedParams{
		Name:   name,
		Url:    url,
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("%w: create feed failed: %w", ErrFatal, parseDBErr(err))
	}

	_, err = s.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		FeedID: feed.ID, UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("%w: follow feed failed: %w", ErrFatal, parseDBErr(err))
	}

	fmt.Printf("Feed created: %s @ %s\n", feed.Name, feed.Url)
	return nil
}

func handlerFeeds(ctx context.Context, s *config.State, cmd command) error {
	feeds, err := s.DB.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("get feeds failed: %w", parseDBErr(err))
	}

	for _, feed := range feeds {
		fmt.Printf("%s - %s @ %s\n", feed.UserName, feed.FeedName, feed.Url)
	}
	return nil
}

func handlerFollow(ctx context.Context, s *config.State, cmd command, user database.User) error {
	url := cmd.args[0]

	feed, err := s.DB.GetFeedByUrl(ctx, url)
	if err != nil {
		return fmt.Errorf("%w: feed missing: %w", ErrFatal, parseDBErr(err))
	}

	_, err = s.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: s.Cfg.CurrentUserID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("%w: create follow failed: %w", ErrFatal, parseDBErr(err))
	}

	fmt.Printf("Started following %s @ %s\n", feed.Name, feed.Url)
	return nil
}

func handlerFollowing(ctx context.Context, s *config.State, cmd command, user database.User) error {
	follows, err := s.DB.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("get follows failed: %w", parseDBErr(err))
	}

	for _, follow := range follows {
		fmt.Printf("- %s @ %s\n", follow.FeedName, follow.FeedUrl)
	}
	return nil
}

func handlerUnfollow(ctx context.Context, s *config.State, cmd command, user database.User) error {
	url := cmd.args[0]
	feed, err := s.DB.GetFeedByUrl(ctx, url)
	if err != nil {
		return fmt.Errorf("%w: Could not find feed with giver URL: %w", ErrFatal, parseDBErr(err))
	}

	err = s.DB.UnfollowFeed(ctx, database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("%w: Could not unfollow feed: %w", ErrFatal, parseDBErr(err))
	}

	fmt.Printf("Stopped following %s @ %s\n", feed.Name, feed.Url)
	return nil
}

type (
	Handler func(ctx context.Context, s *config.State, cmd command) error
	cmdDef  struct {
		ReqArgs int
		Run     Handler
	}
	registry map[string]cmdDef
)

func middlewareLoggedIn(handler func(ctx context.Context, s *config.State, cmd command, user database.User) error) Handler {
	return func(ctx context.Context, s *config.State, cmd command) error {
		user, err := s.DB.GetUserByID(ctx, s.Cfg.CurrentUserID)
		if err != nil {
			return fmt.Errorf("%w: Could not get current user from db: %w", ErrFatal, parseDBErr(err))
		}

		return handler(ctx, s, cmd, user)
	}
}

func NewRegistry() registry {
	registry := registry{
		"register":  {ReqArgs: 1, Run: handlerRegister},
		"login":     {ReqArgs: 1, Run: handlerLogin},
		"reset":     {Run: handlerReset},
		"users":     {Run: handlerUsers},
		"agg":       {Run: handlerAgg},
		"feeds":     {Run: handlerFeeds},
		"addfeed":   {ReqArgs: 2, Run: middlewareLoggedIn(handlerAddFeed)},
		"follow":    {ReqArgs: 1, Run: middlewareLoggedIn(handlerFollow)},
		"following": {Run: middlewareLoggedIn(handlerFollowing)},
		"unfollow":  {ReqArgs: 1, Run: middlewareLoggedIn(handlerUnfollow)},
	}

	return registry
}

func (c registry) Run(ctx context.Context, s *config.State, cmd command) error {
	def, exist := c[cmd.name]
	if !exist {
		return fmt.Errorf("%w: unknown command %s", ErrFatal, cmd.name)
	}

	argLen := len(cmd.args)
	if def.ReqArgs != -1 && argLen != def.ReqArgs {
		return fmt.Errorf("%w: %s requires %d args, got %d", ErrFatal, cmd.name, def.ReqArgs, argLen)
	}
	return def.Run(ctx, s, cmd)
}
