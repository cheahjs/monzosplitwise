package splitwise

import "time"

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	// We don't care about the other attributes for now
}

type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	GroupType string    `json:"group_type"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Expense struct {
	ID              int       `json:"id"`
	GroupID         int       `json:"group_id"`
	FriendshipID    int       `json:"friendship_id"`
	ExpenseBundleID int       `json:"expense_bundle_id"`
	Description     string    `json:"description"`
	Details         string    `json:"details"`
	Payment         bool      `json:"payment"`
	Cost            string    `json:"cost"`
	Date            time.Time `json:"date"`
	CreatedAt       time.Time `json:"created_at"`
	//CreatedBy        string    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	//UpdatedBy        string    `json:"updated_by"`
	DeletedAt time.Time `json:"deleted_at"`
	//DeletedBy        string    `json:"deleted_by"`
}
