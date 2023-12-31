// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package query

// グループチャットリードモデル
//
// NOTE: リードモデルはDTOとして利用されるものです。
// 特段振る舞いのようなものはありません。
type GroupChat struct {
	// グループチャットID
	ID string `json:"id"`
	// グループチャット名
	Name string `json:"name"`
	// 管理者ID
	OwnerID string `json:"ownerId"`
	// 作成日時
	CreatedAt string `json:"createdAt"`
}

// メンバーリードモデル
type Member struct {
	// メンバーID
	ID string `json:"id"`
	// グループチャットID
	GroupChatID string `json:"groupChatId"`
	// アカウントID
	AccountID string `json:"accountId"`
	// ロール
	Role string `json:"role"`
	// 作成日時
	CreatedAt string `json:"createdAt"`
}

// メッセージリードモデル
type Message struct {
	// メッセージID
	ID string `json:"id"`
	// グループチャットID
	GroupChatID string `json:"groupChatId"`
	// アカウントID
	AccountID string `json:"accountId"`
	// メッセージ本文
	Text string `json:"text"`
	// 作成日時
	CreatedAt string `json:"createdAt"`
}
