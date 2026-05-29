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

func NewCommand(args []string) command {
	return command{name: args[1], args: args[2:]}
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

func handlerLogin(s *config.State, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("%w: invalid args. expected 1, got %d", ErrFatal, numArgs)
	}

	userName := cmd.args[0]
	user, err := s.DB.GetUserByName(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("%w: get user failed → %w", ErrFatal, parseDBErr(err))
	}

	s.Cfg.SetUser(user.ID)
	fmt.Printf("User set: %s\n", userName)
	return nil
}

func handlerRegister(s *config.State, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("%w: invalid args. expected 1, got %d", ErrFatal, numArgs)
	}

	userName := cmd.args[0]
	userID := uuid.New()
	_, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		Name: userName,
		ID:   userID,
	})
	if err != nil {
		return fmt.Errorf("%w: register failed → %w", ErrFatal, parseDBErr(err))
	}

	s.Cfg.SetUser(userID)
	fmt.Printf("User created: %s\n", userName)
	return nil
}

func handlerReset(s *config.State, cmd command) error {
	if err := s.DB.ResetUsers(context.Background()); err != nil {
		return fmt.Errorf("%w: reset failed → %w", ErrFatal, parseDBErr(err))
	}

	fmt.Println("Reset successful")
	return nil
}

func handlerUsers(s *config.State, cmd command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("%w: get users failed → %w", ErrFatal, parseDBErr(err))
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

func handlerAgg(s *config.State, cmd command) error {
	feed, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("%w: fetch feed failed → %w", ErrFatal, err)
	}
	fmt.Println(*feed)
	return nil
}

func handlerAddFeed(s *config.State, cmd command) error {
	userID := s.Cfg.CurrentUserID
	if numArgs := len(cmd.args); numArgs != 2 {
		return fmt.Errorf("%w: invalid args. expected 2, got %d", ErrFatal, numArgs)
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		Name:   name,
		Url:    url,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("%w: create feed failed → %w", ErrFatal, parseDBErr(err))
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		FeedID: feed.ID, UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("%w: follow feed failed → %w", ErrFatal, parseDBErr(err))
	}

	fmt.Printf("Feed created: %s @ %s\n", feed.Name, feed.Url)
	return nil
}

func handlerFeeds(s *config.State, cmd command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("get feeds failed → %w", parseDBErr(err))
	}

	for _, feed := range feeds {
		fmt.Printf("%s - %s @ %s\n", feed.UserName, feed.FeedName, feed.Url)
	}
	return nil
}

func handlerFollow(s *config.State, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("%w: invalid args. expected 1, got %d", ErrFatal, numArgs)
	}
	url := cmd.args[0]

	feed, err := s.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("%w: feed missing → %w", ErrFatal, parseDBErr(err))
	}

	follow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: s.Cfg.CurrentUserID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("%w: create follow failed → %w", ErrFatal, parseDBErr(err))
	}

	user, err := s.DB.GetUserByID(context.Background(), s.Cfg.CurrentUserID)
	if err != nil {
		return fmt.Errorf("%w: get user failed → %w", ErrFatal, parseDBErr(err))
	}

	fmt.Printf("%s follows %s\n", user.Name, follow.FeedName)
	return nil
}

func handlerFollowing(s *config.State, cmd command) error {
	follows, err := s.DB.GetFeedFollowsForUser(context.Background(), s.Cfg.CurrentUserID)
	if err != nil {
		return fmt.Errorf("get follows failed → %w", parseDBErr(err))
	}

	for _, follow := range follows {
		fmt.Printf("%s @ %s\n", follow.FeedName, follow.FeedUrl)
	}
	return nil
}

type commands struct {
	registry map[string]func(*config.State, command) error
}

func NewCommands() commands {
	cmds := commands{make(map[string]func(*config.State, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)
	return cmds
}

func (c *commands) Run(s *config.State, cmd command) error {
	handler, exist := c.registry[cmd.name]
	if !exist {
		return fmt.Errorf("%w: unknown command %s", ErrFatal, cmd.name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*config.State, command) error) {
	c.registry[name] = f
}
