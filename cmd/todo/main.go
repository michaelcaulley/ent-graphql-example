package main

import (
	"context"
	"database/sql"
	"entgo.io/ent/privacy"
	"fmt"
	"github.com/rs/zerolog"
	"net/http"
	"os"

	"todo"
	"todo/ent"
	"todo/ent/migrate"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/debug"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"

	_ "github.com/mattn/go-sqlite3"
)

func loadTestData(client *ent.Client) {
	allow := privacy.DecisionContext(context.Background(), privacy.Allow)
	manager := client.User.Create().SetName("Manager One").SaveX(allow)

	one := client.User.Create().SetName("Subordinate One").SetManager(manager).SaveX(allow)
	client.Todo.Create().SetOwner(one).SetText("todo 1a").SaveX(allow)
	client.Todo.Create().SetOwner(one).SetText("todo 1b").SaveX(allow)

	two := client.User.Create().SetName("Subordinate Two").SetManager(manager).SaveX(allow)
	client.Todo.Create().SetOwner(two).SetText("todo 2a").SaveX(allow)
	client.Todo.Create().SetOwner(two).SetText("todo 2b").SaveX(allow)

	three := client.User.Create().SetName("Subordinate Three").SetManager(manager).SaveX(allow)
	client.Todo.Create().SetOwner(three).SetText("todo 3a").SaveX(allow)
	client.Todo.Create().SetOwner(three).SetText("todo 3b").SaveX(allow)

	four := client.User.Create().SetName("Subordinate Four").SetManager(manager).SaveX(allow)
	client.Todo.Create().SetOwner(four).SetText("todo 4a").SaveX(allow)
	client.Todo.Create().SetOwner(four).SetText("todo 4b").SaveX(allow)

	managerTwo := client.User.Create().SetName("Manager Two").SaveX(allow)

	five := client.User.Create().SetName("Subordinate five").SetManager(managerTwo).SaveX(allow)
	client.Todo.Create().SetOwner(five).SetText("todo 5a").SaveX(allow)
	client.Todo.Create().SetOwner(five).SetText("todo 5b").SaveX(allow)

	six := client.User.Create().SetName("Subordinate Six").SetManager(managerTwo).SaveX(allow)
	client.Todo.Create().SetOwner(six).SetText("todo 6a").SaveX(allow)
	client.Todo.Create().SetOwner(six).SetText("todo 6b").SaveX(allow)
}

func main() {
	var cli struct {
		Addr  string `name:"address" default:":8081" help:"Address to listen on."`
		Debug bool   `name:"debug" help:"Enable debugging mode."`
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	kong.Parse(&cli)
	dbURL := "file:ent?mode=memory&cache=shared&_fk=1"
	db, _ := sql.Open(dialect.SQLite, dbURL)
	db = sqldblogger.OpenDriver(dbURL, db.Driver(), zerologadapter.New(log.Logger))
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	if err := client.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
	); err != nil {
		log.Fatal().Err(err).Msg("running schema migration")

	}

	loadTestData(client)

	srv := handler.NewDefaultServer(todo.NewSchema(client))
	srv.Use(entgql.Transactioner{TxOpener: client})
	if cli.Debug {
		srv.Use(&debug.Tracer{})
	}
	http.Handle("/",
		playground.Handler("Todo", "/query"),
	)
	http.Handle("/query", srv)
	log.Info().Msg(fmt.Sprintf("listening on %v", cli.Addr))
	if err := http.ListenAndServe(cli.Addr, nil); err != nil {
		log.Fatal().Err(err).Msg("http server terminated")
	}
}
