package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Account Interface
type Account interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	CheckBalance() float64
	GetHistory() []string
}

// Savings Account Struct
type SavingsAccount struct {
	Name    string
	Balance *float64
	Limit   float64
	History []string
}

// Current Account Struct
type CurrentAccount struct {
	Name      string
	Balance   *float64
	Overdraft float64
	History   []string
}

// Global Account Storage
var (
	accounts = make(map[string]Account)
	mutex    sync.Mutex
)

// Deposit Money
func (a *SavingsAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("Invalid deposit amount")
	}
	*a.Balance += amount
	a.History = append(a.History, fmt.Sprintf("Deposited: $%.2f", amount))
	return nil
}

func (a *CurrentAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("Invalid deposit amount")
	}
	*a.Balance += amount
	a.History = append(a.History, fmt.Sprintf("Deposited: $%.2f", amount))
	return nil
}

// Withdraw Money
func (a *SavingsAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("Invalid withdrawal amount")
	}
	if amount > *a.Balance {
		return fmt.Errorf("Insufficient funds! Available balance: $%.2f", *a.Balance)
	}
	if (*a.Balance - amount) < a.Limit {
		return fmt.Errorf("Transaction denied! You must maintain a minimum balance of $%.2f", a.Limit)
	}

	*a.Balance -= amount
	transaction := fmt.Sprintf("Withdrew: $%.2f, Final Balance: $%.2f", amount, *a.Balance)
	a.History = append(a.History, transaction)

	return nil
}

func (a *CurrentAccount) Withdraw(amount float64) error {
	if *a.Balance-amount < -a.Overdraft {
		return fmt.Errorf("Overdraft limit exceeded!")
	}
	*a.Balance -= amount
	a.History = append(a.History, fmt.Sprintf("Withdrew: $%.2f", amount))
	return nil
}

// Check Balance
func (a *SavingsAccount) CheckBalance() float64 { return *a.Balance }
func (a *CurrentAccount) CheckBalance() float64 { return *a.Balance }

// Get Transaction History
func (a *SavingsAccount) GetHistory() []string { return a.History }
func (a *CurrentAccount) GetHistory() []string { return a.History }

// Create Account Handler
func createAccount(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	balance, _ := strconv.ParseFloat(r.FormValue("balance"), 64)
	accountType := r.FormValue("accountType")

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := accounts[name]; exists {
		sendResponse(w, "Account already exists")
		return
	}

	initialBalance := balance
	if accountType == "Savings" {
		accounts[name] = &SavingsAccount{Name: name, Balance: &initialBalance, Limit: 500.00}
	} else {
		accounts[name] = &CurrentAccount{Name: name, Balance: &initialBalance, Overdraft: 1000.00}
	}

	sendResponse(w, "Account created successfully")
}

// Deposit Money Handler
func depositMoney(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

	mutex.Lock()
	defer mutex.Unlock()

	if acc, exists := accounts[name]; exists {
		err := acc.Deposit(amount)
		if err != nil {
			sendResponse(w, err.Error())
		} else {
			sendResponse(w, fmt.Sprintf("Deposit successful! New Balance: $%.2f", acc.CheckBalance()))
		}
	} else {
		sendResponse(w, "Account not found")
	}
}

// Withdraw Money Handler
func withdrawMoney(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

	mutex.Lock()
	defer mutex.Unlock()

	if acc, exists := accounts[name]; exists {
		err := acc.Withdraw(amount)
		if err != nil {
			sendResponse(w, err.Error())
		} else {
			sendResponse(w, fmt.Sprintf("Withdrawal successful! New Balance: $%.2f", acc.CheckBalance()))
		}
	} else {
		sendResponse(w, "Account not found")
	}
}

// Check Balance Handler
func checkBalance(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	mutex.Lock()
	defer mutex.Unlock()

	if acc, exists := accounts[name]; exists {
		sendResponse(w, fmt.Sprintf("Balance: $%.2f", acc.CheckBalance()))
	} else {
		sendResponse(w, "Account not found")
	}
}

// Transaction History Handler
func transactionHistory(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	mutex.Lock()
	defer mutex.Unlock()

	if acc, exists := accounts[name]; exists {
		history := acc.GetHistory()
		response := ""
		for _, transaction := range history {
			response += transaction + "<br>"
		}
		sendResponse(w, response)
	} else {
		sendResponse(w, "Account not found")
	}
}

// Send Response Helper Function
func sendResponse(w http.ResponseWriter, message string) {
	w.Write([]byte(message))
}

// Home Page with UI
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Banking System</title>
		<style>
			body { font-family: Arial, sans-serif; background-color: #f4f4f4; text-align: center; padding: 20px; }
			h2 { color: #007bff; }
			form {
				background: white; padding: 15px; margin: 10px auto; border-radius: 8px;
				box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); width: 300px;
			}
			input, select, button { margin: 10px; padding: 8px; width: 90%; border-radius: 5px; }
			button { background-color: #007bff; color: white; border: none; cursor: pointer; }
			button:hover { background-color: #0056b3; }
			#resultModal {
				display: none; position: fixed; left: 50%; top: 50%; transform: translate(-50%, -50%);
				background: white; padding: 20px; border-radius: 10px; box-shadow: 0 4px 8px rgba(0,0,0,0.2);
			}
		</style>
	</head>
	<body>
		<h2>🏦 Banking System</h2>

		<!-- Create Account Form -->
		<form onsubmit="return handleSubmit(event, '/create')">
			<h3>Create Account</h3>
			<input type="text" name="name" placeholder="Name" required>
			<input type="number" name="balance" placeholder="Initial Balance" required>
			<select name="accountType">
				<option value="Savings">Savings</option>
				<option value="Current">Current</option>
			</select>
			<button type="submit">Create Account</button>
		</form>

		<!-- Deposit Money Form -->
		<form onsubmit="return handleSubmit(event, '/deposit')">
			<h3>Deposit Money</h3>
			<input type="text" name="name" placeholder="Name" required>
			<input type="number" name="amount" placeholder="Amount" required>
			<button type="submit">Deposit</button>
		</form>

		<!-- Withdraw Money Form -->
		<form onsubmit="return handleSubmit(event, '/withdraw')">
			<h3>Withdraw Money</h3>
			<input type="text" name="name" placeholder="Name" required>
			<input type="number" name="amount" placeholder="Amount" required>
			<button type="submit">Withdraw</button>
		</form>

		<!-- Check Balance Form -->
		<form onsubmit="return checkBalance(event)">
			<h3>Check Balance</h3>
			<input type="text" id="balanceName" placeholder="Enter Account Name" required>
			<button type="submit">Check Balance</button>
		</form>

		<!-- Transaction History Form -->
		<form onsubmit="return fetchHistory(event)">
			<h3>Transaction History</h3>
			<input type="text" id="historyName" placeholder="Enter Account Name" required>
			<button type="submit">View History</button>
		</form>

		<!-- Modal for Showing Responses -->
		<div id="resultModal"></div>

		<script>
			function handleSubmit(event, url) {
				event.preventDefault();
				const formData = new FormData(event.target);
				fetch(url, { method: 'POST', body: formData })
					.then(res => res.text())
					.then(data => showModal(data));
			}

			function checkBalance(event) {
				event.preventDefault();
				const name = document.getElementById('balanceName').value;
				fetch('/balance?name=' + name)
					.then(res => res.text())
					.then(data => showModal(data));
			}

			function fetchHistory(event) {
				event.preventDefault();
				const name = document.getElementById('historyName').value;
				fetch('/history?name=' + name)
					.then(res => res.text())
					.then(data => showModal(data));
			}

			function showModal(message) {
				const modal = document.getElementById('resultModal');
				modal.innerHTML = message;
				modal.style.display = 'block';
				setTimeout(() => modal.style.display = 'none', 5000); // Hide modal after 5 seconds
			}
		</script>
	</body>
	</html>`
	w.Write([]byte(tmpl))
}

func main() {
	http.HandleFunc("/", homeHandler)               // Home Page with UI
	http.HandleFunc("/create", createAccount)       // Create Account
	http.HandleFunc("/deposit", depositMoney)       // Deposit Money
	http.HandleFunc("/withdraw", withdrawMoney)     // Withdraw Money
	http.HandleFunc("/balance", checkBalance)       // Check Balance
	http.HandleFunc("/history", transactionHistory) // Transaction History

	fmt.Println("Server is running on port 8080...")
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
