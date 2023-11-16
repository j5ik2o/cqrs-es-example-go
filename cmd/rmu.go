package cmd

import (
	"cqrs-es-example-go/pkg/rmu"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"

	"github.com/spf13/cobra"
)

// rmuCmd represents the rmu command
var rmuCmd = &cobra.Command{
	Use:   "rmu",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbUrl := env.String("", "DATABASE_URL")
		dataSourceName := fmt.Sprintf("%s?parseTime=true", dbUrl)
		db, err := sqlx.Connect("mysql", dataSourceName)
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
		dao := rmu.NewGroupChatDaoImpl(db)
		readModelUpdater := rmu.NewReadModelUpdater(dao)
		lambda.Start(readModelUpdater.UpdateReadModel)
	},
}

func init() {
	rootCmd.AddCommand(rmuCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rmuCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rmuCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
