package example2chattemplate

import (
	"context"
	"testing"
)

func TestChatTemplate(t *testing.T) {

	t.Run("ChatTemplate", func(t *testing.T) {
		ctx := context.Background()
		ChatTemplate(ctx)
	})

}
