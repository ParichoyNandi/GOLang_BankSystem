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
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Banking System</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				background: url('/static/bg.jpg') no-repeat center center fixed;
				background-size: cover;
				color: white;
				text-align: center;
				margin: 0;
				padding: 0;
			}
			h2 { margin-top: 20px; }
			.container {
				display: flex;
				flex-direction: column;
				align-items: center;
				margin-top: 20px;
			}
			.card {
				background: rgba(255, 255, 255, 0.2); /* Translucent effect */
				color: white;
				padding: 15px;
				border-radius: 10px;
				box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
				width: 350px;
				margin: 10px;
				transition: all 0.3s ease-in-out;
				backdrop-filter: blur(10px); /* Glassmorphism effect */
			}
			.card-header {
				background: rgba(0, 123, 255, 0.7);
				color: white;
				padding: 10px;
				cursor: pointer;
				border-radius: 10px 10px 0 0;
				font-weight: bold;
			}
			.card-content {
				display: none;
				padding: 15px;
			}
			.card.active .card-content {
				display: block;
			}
			.card.active .card-header {
				background: rgba(1, 33, 66, 0.8);
			}
			input, select, button {
				margin: 8px 0;
				padding: 10px;
				width: 95%;
				border-radius: 5px;
				border: 1px solid #ddd;
			}
			button {
				background-color: rgba(45, 220, 255, 0.8);
				color: white;
				border: none;
				cursor: pointer;
			}
			button:hover { background-color: rgba(0, 86, 179, 0.9); }
			#resultModal {
				display: none;
				position: fixed;
				left: 50%;
				top: 50%;
				transform: translate(-50%, -50%);
				background: rgba(255, 255, 255, 0.9);
				color: black;
				padding: 20px;
				border-radius: 10px;
				box-shadow: 0 4px 8px rgba(0,0,0,0.2);
				backdrop-filter: blur(5px);
			}
			@media (max-width: 768px) {
				.container { width: 100%; }
				.card { width: 90%; }
			}
		</style>
	</head>
	<body>
		<h2>🏦 Welcome to Your Bank</h2>
		<div class="container">

			<!-- Create Account -->
			<div class="card">
				<div class="card-header" onclick="toggleCard(this)">🆕 Create Account ⬇️</div>
				<div class="card-content">
					<form onsubmit="return handleSubmit(event, '/create')">
						<input type="text" name="name" placeholder="Full Name" required>
						<input type="number" name="balance" placeholder="Initial Balance" required>
						<select name="accountType">
							<option value="Savings">Savings</option>
							<option value="Current">Current</option>
						</select>
						<button type="submit">Create</button>
					</form>
				</div>
			</div>

			<!-- Deposit Money -->
			<div class="card">
				<div class="card-header" onclick="toggleCard(this)">💰 Deposit Money ⬇️</div>
				<div class="card-content">
					<form onsubmit="return handleSubmit(event, '/deposit')">
						<input type="text" name="name" placeholder="Account Name" required>
						<input type="number" name="amount" placeholder="Amount" required>
						<button type="submit">Deposit</button>
					</form>
				</div>
			</div>

			<!-- Withdraw Money -->
			<div class="card">
				<div class="card-header" onclick="toggleCard(this)">🏧 Withdraw Money ⬇️</div>
				<div class="card-content">
					<form onsubmit="return handleSubmit(event, '/withdraw')">
						<input type="text" name="name" placeholder="Account Name" required>
						<input type="number" name="amount" placeholder="Amount" required>
						<button type="submit">Withdraw</button>
					</form>
				</div>
			</div>

			<!-- Check Balance -->
			<div class="card">
				<div class="card-header" onclick="toggleCard(this)">📊 Check Balance ⬇️</div>
				<div class="card-content">
					<form onsubmit="return checkBalance(event)">
						<input type="text" id="balanceName" placeholder="Account Name" required>
						<button type="submit">Check</button>
					</form>
				</div>
			</div>

			<!-- Transaction History -->
			<div class="card">
				<div class="card-header" onclick="toggleCard(this)">📜 Transaction History ⬇️</div>
				<div class="card-content">
					<form onsubmit="return fetchHistory(event)">
						<input type="text" id="historyName" placeholder="Account Name" required>
						<button type="submit">View</button>
					</form>
				</div>
			</div>

		</div>

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
					.then(res => res.json())  
					.then(data => showHistoryModal(data));
			}

			function showModal(message) {
				const modal = document.getElementById('resultModal');
				modal.innerHTML = '<h3>📢 Notification</h3><p>' + message + '</p>';
				modal.style.display = 'block';
				setTimeout(() => modal.style.display = 'none', 5000);
			}

			function toggleCard(header) {
				const card = header.parentElement;
				card.classList.toggle('active');
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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Server is running on port 8080...")
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
