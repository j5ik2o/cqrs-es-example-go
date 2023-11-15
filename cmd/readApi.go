package cmd

import (
	"cqrs-es-example-go/query/graph"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

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
		port := os.Getenv("PORT")
		if port == "" {
			port = defaultPort
		}

		srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		http.Handle("/query", srv)

		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	},
}

const defaultPort = "8080"

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
