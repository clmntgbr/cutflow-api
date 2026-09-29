package command

import (
	"context"
	"fmt"

	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/storage"

	"github.com/spf13/cobra"
)

func NewPurgeStorageCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "purge-storage",
		Short: "Delete all objects in the configured S3/MinIO bucket",
		Long:  "Removes every object from STORAGE_BUCKET. Does not drop the bucket. Requires --yes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("refusing to purge storage without --yes")
			}

			env := config.Load()
			store, err := storage.NewMinIOStorage(env)
			if err != nil {
				return fmt.Errorf("storage client: %w", err)
			}

			deleted, err := store.DeleteAllObjects(context.Background())
			if err != nil {
				return err
			}

			fmt.Printf("Purged %d object(s) from bucket %s\n", deleted, env.StorageBucket)
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm destructive storage purge")
	return cmd
}
