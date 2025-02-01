# File-Based Database in Golang

## Overview
This project is a simple file-based database written in Golang that allows storing, reading, and deleting JSON records efficiently. It uses a directory-based approach where collections are stored as directories and individual records as JSON files.

## Features
- Store data as JSON files within structured directories.
- Support for reading, writing, and deleting records.
- Thread-safe operations using mutex locks.
- Custom logging for better debugging.
- Minimal dependencies for lightweight execution.

## Installation
Ensure you have Go installed on your system. Then, clone this repository:
```sh
git clone https://github.com/asmit990/ruktxplorer.git
cd ruktxplorer
```

## Usage

### Initialize Database
Create a database instance by specifying a directory path:
```go
logger := lumber.NewConsoleLogger(lumber.DEBUG)
db, err := New("./database", &Options{Logger: logger})
if err != nil {
    fmt.Println("Error initializing database:", err)
    return
}
```

### Write Data
Store a user record in the database:
```go
user := User{
    Name: "John",
    Age: "23",
    Contact: "9354074216",
    Company: "RUKTIFY",
    Address: Address{
        City: "Bangalore",
        State: "Karnataka",
        Country: "India",
        Pincode: "42019",
    },
}

db.Write("users", user.Name, user)
```

### Read Data
Retrieve a user record from the database:
```go
var retrievedUser User
db.Read("users", "John", &retrievedUser)
fmt.Println("Retrieved User:", retrievedUser)
```

### Read All Records
Retrieve all users in the collection:
```go
records, err := db.ReadAll("users")
if err != nil {
    fmt.Println("ReadAll Error:", err)
    return
}
fmt.Println("All Users:", records)
```

### Delete Data
Delete a specific user record:
```go
db.Delete("users", "John")
```

### Delete Collection
Delete all user records:
```go
db.Delete("users", "")
```

## Dependencies
- `github.com/jcelliott/lumber` (For logging)

Install dependencies using:
```sh
go mod tidy
```

## License
This project is licensed under the MIT License.

