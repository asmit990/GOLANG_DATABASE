package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type Logger interface {
	Fatal(string, ...interface{})
	Error(string, ...interface{})
	Warn(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Trace(string, ...interface{})
}

type Driver struct {
	mutex   sync.Mutex
	mutexes map[string]*sync.Mutex
	dir     string
	log     Logger
}

type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}
	if options != nil {
		opts = *options
	}
	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating the database at '%s' ...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)
}
func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to save records")
	}
	if resource == "" {
		return fmt.Errorf("missing resource - unable to save record (no name)!")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	finalPath := filepath.Join(dir, resource+".json")
	tmpPath := finalPath + ".tmp"

	d.log.Debug("Creating directory: %s", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		d.log.Error("Failed to create directory: %v", err)
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		d.log.Error("JSON Marshalling failed: %v", err)
		return err
	}
	b = append(b, byte('\n'))

	d.log.Debug("Writing to temp file: %s", tmpPath)
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		d.log.Error("Failed to write temp file: %v", err)
		return err
	}

	d.log.Debug("Renaming temp file to final: %s", finalPath)
	return os.Rename(tmpPath, finalPath)
}


func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - unable to read")
	}
	if resource == "" {
		return fmt.Errorf("missing resource - unable to read (no name)")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := os.ReadFile(record + ".json")
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("missing collection - unable to read")
	}

	dir := filepath.Join(d.dir, collection)
	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var records []string
	for _, file := range files {
		b, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		records = append(records, string(b))
	}

	return records, nil
}

func (d *Driver) Delete(collection, resource string) error {
	if collection == "" {
		return fmt.Errorf("missing collection - unable to delete")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	path := filepath.Join(d.dir, collection, resource)

	fi, err := stat(path)
	if err != nil {
		return fmt.Errorf("unable to find file or directory named %v\n", path)
	}

	if fi.Mode().IsDir() {
		return os.RemoveAll(path)
	}
	if fi.Mode().IsRegular() {
		return os.Remove(path + ".json")
	}
	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}

func stat(path string) (os.FileInfo, error) {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return fi, err
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	// Get absolute path for better debugging
	dir, err := filepath.Abs("./database")
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return
	}

	// Create a custom logger to see what's happening
	logger := lumber.NewConsoleLogger(lumber.DEBUG)
	
	db, err := New(dir, &Options{Logger: logger})
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	logger.Debug("Database initialized at: %s", dir)

	employees := []User{
		{"John", "23", "9354074216", "RUKTIFY", Address{"Bangalore", "Karnataka", "India", "42019"}},
		{"Alice", "29", "8789674123", "TechFlow", Address{"San Francisco", "California", "USA", "94105"}},
		{"Bob", "35", "9078563412", "DataCorp", Address{"New York", "New York", "USA", "10001"}},
	}

	// Create users collection directory explicitly
	usersDir := filepath.Join(dir, "users")
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		fmt.Println("Error creating users directory:", err)
		return
	}

	logger.Debug("Users directory created at: %s", usersDir)

	for _, value := range employees {
		logger.Debug("Writing user: %s", value.Name)
		if err := db.Write("users", value.Name, value); err != nil {
			fmt.Println("Write Error for user", value.Name, ":", err)
		} else {
			logger.Debug("Successfully wrote user: %s", value.Name)
		}
	}

	// Verify files were created
	files, err := os.ReadDir(usersDir)
	if err != nil {
		fmt.Println("Error reading users directory:", err)
		return
	}

	logger.Debug("Files in users directory:")
	for _, file := range files {
		logger.Debug("- %s", file.Name())
	}

	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("ReadAll Error:", err)
		return
	}
	
	if len(records) == 0 {
		logger.Warn("No records found in the users collection")
	} else {
		logger.Info("Found %d records", len(records))
	}

	fmt.Println("Raw Records:", records)

	var allUsers []User
	for _, record := range records {
		var user User
		if err := json.Unmarshal([]byte(record), &user); err != nil {
			fmt.Println("JSON Unmarshal Error:", err)
			continue
		}
		allUsers = append(allUsers, user)
	}
	fmt.Println("All Users:", allUsers)

	// Example of deleting a specific user
	// if err := db.Delete("users", "John"); err != nil {
	// 	fmt.Println("Delete Error:", err)
	// }

	// Example of deleting the entire collection
// 	if err := db.Delete("users", ""); err != nil {
// 		fmt.Println("Delete Collection Error:", err)
	}
