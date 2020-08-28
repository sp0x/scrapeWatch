package scrapeWatch

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

func BindConfig() {
	viper.AutomaticEnv()
	_ = viper.BindEnv("firebase_project")
	_ = viper.BindEnv("firebase_credentials_file")
}

const schemeTopic = "scrapescheme"
const schemeErrorTopic = "scheme.topic"

type FirebaseConfig struct {
	Project     string
	Credentials string
}

func GetFirebaseConfig() (*FirebaseConfig, error) {
	project := viper.GetString("GOOGLE_CLOUD_PROJECT")
	creds := viper.GetString("GOOGLE_APPLICATION_CREDENTIALS")
	if creds == "" {
		return nil, errors.New("no firebase credentials found")
	}
	return &FirebaseConfig{
		Project:     project,
		Credentials: creds,
	}, nil
}

func NewFirebaseFromEnv() (*firestore.Client, error) {
	c, err := GetFirebaseConfig()
	if err != nil {
		return nil, err
	}
	return NewFirebase(c.Project, c.Credentials)
}

func NewFirebase(projectId string, credentialsFile string) (*firestore.Client, error) {
	ctx := context.Background()
	var options []option.ClientOption
	if credentialsFile != "" {
		options = append(options, option.WithCredentialsFile(credentialsFile))
	}
	// credentials file option is optional, by default it will use GOOGLE_APPLICATION_CREDENTIALS
	// environment variable, this is a default method to connect to Google services
	client, err := firestore.NewClient(ctx, projectId, options...)
	return client, err
}

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

var initialized = false
var firebase *firestore.Client

func initialize() {
	if initialized {
		return
	}
	initialized = true
	BindConfig()
	fb, err := NewFirebaseFromEnv()
	if err != nil {
		fmt.Printf("error initializing firestore: %v", err)
		//os.Exit(1)
	}
	firebase = fb
}

// HelloPubSub consumes a Pub/Sub message.
func NonErrorStatusReceived(ctx context.Context, m PubSubMessage) error {
	name := string(m.Data) // Automatically decoded from base64.
	initialize()
	fmt.Printf(name)
	return nil
}
