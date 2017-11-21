package splitwise

import "time"

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Picture   struct {
		Small  string `json:"small"`
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"picture"`
	Email              string      `json:"email"`
	RegistrationStatus string      `json:"registration_status"`
	ForceRefreshAt     interface{} `json:"force_refresh_at"`
	Locale             string      `json:"locale"`
	CountryCode        string      `json:"country_code"`
	DateFormat         string      `json:"date_format"`
	DefaultCurrency    string      `json:"default_currency"`
	DefaultGroupID     int         `json:"default_group_id"`
	NotificationsRead  time.Time   `json:"notifications_read"`
	NotificationsCount int         `json:"notifications_count"`
	Notifications      struct {
		AddedAsFriend  bool `json:"added_as_friend"`
		AddedToGroup   bool `json:"added_to_group"`
		ExpenseAdded   bool `json:"expense_added"`
		ExpenseUpdated bool `json:"expense_updated"`
		Bills          bool `json:"bills"`
		Payments       bool `json:"payments"`
		MonthlySummary bool `json:"monthly_summary"`
		Announcements  bool `json:"announcements"`
	} `json:"notifications"`
}

type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
	Members   []struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Picture   struct {
			Medium string `json:"medium"`
		} `json:"picture"`
		// Balance []struct {
		// 	Amount       string `json:"amount"`
		// 	CurrencyCode string `json:"currency_code"`
		// } `json:"balance"`
	} `json:"members"`
	SimplifyByDefault bool          `json:"simplify_by_default"`
	OriginalDebts     []interface{} `json:"original_debts"`
	SimplifiedDebts   []interface{} `json:"simplified_debts"`
	Whiteboard        interface{}   `json:"whiteboard,omitempty"`
	GroupType         string        `json:"group_type,omitempty"`
	InviteLink        string        `json:"invite_link,omitempty"`
}

type Expense struct {
	ID                     int         `json:"id"`
	GroupID                int         `json:"group_id"`
	FriendshipID           interface{} `json:"friendship_id"`
	ExpenseBundleID        int         `json:"expense_bundle_id"`
	Description            string      `json:"description"`
	Repeats                bool        `json:"repeats"`
	RepeatInterval         string      `json:"repeat_interval"`
	EmailReminder          bool        `json:"email_reminder"`
	EmailReminderInAdvance int         `json:"email_reminder_in_advance"`
	NextRepeat             interface{} `json:"next_repeat"`
	Details                string      `json:"details"`
	CommentsCount          int         `json:"comments_count"`
	Payment                bool        `json:"payment"`
	CreationMethod         string      `json:"creation_method"`
	TransactionMethod      string      `json:"transaction_method"`
	TransactionConfirmed   bool        `json:"transaction_confirmed"`
	TransactionID          interface{} `json:"transaction_id"`
	Cost                   string      `json:"cost"`
	CurrencyCode           string      `json:"currency_code"`
	Repayments             []struct {
		From   int    `json:"from"`
		To     int    `json:"to"`
		Amount string `json:"amount"`
	} `json:"repayments"`
	Date      time.Time   `json:"date"`
	CreatedAt time.Time   `json:"created_at"`
	CreatedBy User        `json:"created_by"`
	UpdatedAt time.Time   `json:"updated_at"`
	UpdatedBy interface{} `json:"updated_by"`
	DeletedAt interface{} `json:"deleted_at"`
	DeletedBy interface{} `json:"deleted_by"`
	Category  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Receipt struct {
		Large    interface{} `json:"large"`
		Original interface{} `json:"original"`
	} `json:"receipt"`
	Users []struct {
		User       User   `json:"user"`
		UserID     int    `json:"user_id"`
		PaidShare  string `json:"paid_share"`
		OwedShare  string `json:"owed_share"`
		NetBalance string `json:"net_balance"`
	} `json:"users"`
}
