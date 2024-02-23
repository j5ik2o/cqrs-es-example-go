package cmd

import (
	"cqrs-es-example-go/pkg/query/graph"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/spf13/cobra"
	"log/slog"
	"net/http"
	"os"
)

const readApiDefaultPort = 28082

// readApiCmd represents the readApi command
var readApiCmd = &cobra.Command{
	Use:   "readApi",
	Short: "Read API",
	Long:  "Read API",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		apiPort := env.Int(readApiDefaultPort, "API_PORT")
		apiHost := env.String("0.0.0.0", "API_HOST")
		dbUrl := env.String("", "DATABASE_URL")

		if dbUrl == "" {
			panic("DATABASE_URL is required")
		}

		slog.Info(fmt.Sprintf("apiPort = %v", apiPort))
		slog.Info(fmt.Sprintf("apiHost = %v", apiHost))
		slog.Info(fmt.Sprintf("dbUrl = %v", dbUrl))

		db, err := sqlx.Connect("mysql", fmt.Sprintf("%s?parseTime=true", dbUrl))
		if err != nil {
			panic(err.Error())
		}
		defer func(db *sqlx.DB) {
			if db != nil {
				err := db.Close()
				if err != nil {
					panic(err.Error())
				}
			}
		}(db)

		srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(db)}))

		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		http.Handle("/query", srv)

		endpoint := fmt.Sprintf("%s:%d", apiHost, apiPort)
		slog.Info(fmt.Sprintf("connect to http://%s/ for GraphQL playground", endpoint))
		err = http.ListenAndServe(endpoint, nil)
		if err != nil {
			slog.Error("failed to start server", "error", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(readApiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readApiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readApiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
