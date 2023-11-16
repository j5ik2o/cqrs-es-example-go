package cmd

import (
	"cqrs-es-example-go/pkg/query/graph"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

const defaultPort = "8080"

// readApiCmd represents the readApi command
var readApiCmd = &cobra.Command{
	Use:   "readApi",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiPort := env.Int(8080, "API_PORT")
		apiHost := env.String("0.0.0.0", "API_HOST")
		dbUrl := env.String("", "DATABASE_URL")

		db, err := sqlx.Connect("mysql", fmt.Sprintf("%s?parseTime=true", dbUrl))
		defer func(db *sqlx.DB) {
			if db != nil {
				err := db.Close()
				if err != nil {
					panic(err.Error())
				}
			}
		}(db)

		if err != nil {
			panic(err.Error())
		}
		srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(db)}))

		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		http.Handle("/query", srv)

		endpoint := fmt.Sprintf("%s:%d", apiHost, apiPort)
		log.Printf("connect to http://%s/ for GraphQL playground", endpoint)
		log.Fatal(http.ListenAndServe(endpoint, nil))
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
