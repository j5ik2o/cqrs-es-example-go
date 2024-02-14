package cmd

import (
	"cqrs-es-example-go/pkg/rmu"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// rmuCmd represents the rmu command
var rmuCmd = &cobra.Command{
	Use:   "rmu",
	Short: "Read Model Updater",
	Long:  "Read Model Updater",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		dbUrl := env.String("", "DATABASE_URL")
		if dbUrl == "" {
			panic("DATABASE_URL is required")
		}

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
		readModelUpdater := rmu.NewReadModelUpdater(&dao)
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
