package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/charm"
	"github.com/charmbracelet/charm/ui"
	"github.com/charmbracelet/tea"
)

type TermLinkHandler struct{}

func (th *TermLinkHandler) TokenCreated(l *charm.Link) {
	fmt.Printf("To link a machine, run: \n\n> charm link %s\n", l.Token)
}

func (th *TermLinkHandler) TokenSent(l *charm.Link) {
	fmt.Println("Linking...")
}

func (th *TermLinkHandler) ValidToken(l *charm.Link) {
	fmt.Println("Valid token")
}

func (th *TermLinkHandler) InvalidToken(l *charm.Link) {
	fmt.Println("That token looks invalid.")
}

func (th *TermLinkHandler) Request(l *charm.Link) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Does this look right? (yes/no)\n\n%s\nIP: %s\n", l.RequestPubKey, l.RequestAddr)
	conf, _ := reader.ReadString('\n')
	if strings.ToLower(conf) == "yes\n" {
		return true
	}
	return false
}

func (th *TermLinkHandler) RequestDenied(l *charm.Link) {
	fmt.Println("Not Linked :(")
}

func (th *TermLinkHandler) SameAccount(l *charm.Link) {
	fmt.Println("Linked! You already linked this key btw.")
}

func (th *TermLinkHandler) Success(l *charm.Link) {
	fmt.Println("Linked!")
}

func (th *TermLinkHandler) Timeout(l *charm.Link) {
	fmt.Println("Timed out. Sorry.")
}

func (th *TermLinkHandler) Error(l *charm.Link) {
	fmt.Println("Error, something's wrong.")
}

func main() {
	i := flag.String("i", "", "identity file (ssh key) path")
	flag.Parse()
	cfg, err := charm.ConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if *i != "" {
		cfg.SSHKeyPath = *i
		cfg.ForceKey = true
	}
	cc, err := charm.NewClient(cfg)
	if err == charm.ErrMissingSSHAuth {
		log.Fatal("Missing ssh key. Run `ssh-keygen` to make one or set the `CHARM_SSH_KEY_PATH` env var to your private key path.")
	}
	if err != nil {
		log.Fatal(err)
	}

	tea.UseSysLog("charm-tea")
	if err := ui.NewProgram(cc).Start(); err != nil {
		log.Fatal(err)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}
	switch args[0] {
	case "name":
		if len(args) != 2 {
			log.Fatal("Usage: charm name USERNAME")
		}
		n := args[1]
		u, err := cc.SetName(n)
		if err == charm.ErrNameTaken {
			fmt.Printf("User name '%s' is already taken. Try another!\n", n)
			os.Exit(1)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("@%s ID: %s\n", u.Name, u.CharmID)
	case "jwt":
		jwt, err := cc.JWT()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", jwt)
	case "id":
		id, err := cc.ID()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", id)
	case "bio":
		u, err := cc.Bio()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v", u)
	case "keys":
		ak, err := cc.AuthorizedKeys()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", ak)
	case "link":
		lh := &TermLinkHandler{}
		switch len(args) {
		case 1:
			err := cc.LinkGen(lh)
			if err != nil {
				log.Fatal(err)
			}
		case 2:
			err := cc.Link(lh, args[1])
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("Bad link command")
		}
	default:
		fmt.Printf("'%s' is not a valid command", args[0])
		os.Exit(1)
	}
}
