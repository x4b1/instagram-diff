package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"

	"golang.org/x/term"

	"github.com/xabi93/instagram-diff/instagram"
	"github.com/xabi93/instagram-diff/log"
	"github.com/xabi93/instagram-diff/server"
)

func main() {
	var a App

	a.Init()
	a.Run()
}

type Conf struct {
	sessionFile string
	port        string
}

type App struct {
	cfg Conf
	l   log.Logger
}

func (a *App) Init() {
	a.l = log.NewLogger()
	home, err := os.UserHomeDir()
	if err != nil {
		a.l.Fatal(err.Error())
	}
	flag.StringVar(&a.cfg.sessionFile, "sessionFile", fmt.Sprintf("%s/.instadiff", home), "Session file")

	flag.StringVar(&a.cfg.port, "port", "3000", "Port to expose result")

	flag.Parse()
}

func (a App) Run() {
	i, err := a.login()
	if err != nil {
		a.l.Fatal(err.Error())
	}

	if err := server.Serve(a.cfg.port, i); err != nil {
		a.l.Fatal(err.Error())
	}
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

var cli *instagram.Instagram

func (a App) login() (*instagram.Instadiff, error) {
	var err error
	cli, err = a.restore()
	if err != nil {
		return nil, err
	}
	if cli != nil {
		return instagram.New(cli), nil
	}

	user, pass, err := a.askUserPass()

	fmt.Println("Login...")

	cli, err = instagram.Login(user, pass, a.cfg.sessionFile)
	if err != nil {
		return nil, err
	}

	return instagram.New(cli), nil
}

func (a App) restore() (*instagram.Instagram, error) {
	if _, err := os.Stat(a.cfg.sessionFile); os.IsNotExist(err) {
		return nil, nil
	}

	a.l.Info(fmt.Sprintf("Restoring session from %s", a.cfg.sessionFile))

	i, err := instagram.RestoreSession(a.cfg.sessionFile)
	if err != nil {
		return nil, err
	}

	err = i.Ping()
	if err == nil {
		return i, nil
	}

	if errors.As(err, &instagram.AuthError{}) {
		os.Remove(a.cfg.sessionFile)
		return nil, nil
	}

	return i, nil
}

func (App) askUserPass() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	fmt.Print("\n")

	return username, string(bytePassword), nil
}
