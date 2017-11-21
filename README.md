# MonzoSplitwise
Automatically adding Monzo transactions to Splitwise.

Application searches Monzo transaction history for transactions with notes that contain `#splitwise` or `#splitwise-<groupname>`. `<groupname>` corresponds to the name of a Splitwise group, minus any spaces in the name. If no group is specified, the expense is added to Non-group expenses.

![Screenshot](assets/screenshot.png)

---

# Usage

`go run app/main.go`

Requires:

* [Monzo OAuth client details](https://developers.monzo.com/apps)
  * Currently, the app assumes that the client is a confidential client and has access to refresh tokens.
* [Splitwise OAuth client details](https://secure.splitwise.com/oauth_clients)

Copy `config.json.example` to `config.json`, and fill in the necessary details. Upon first run, the app will guide you through obtaining access tokens for both Monzo and Splitwise.