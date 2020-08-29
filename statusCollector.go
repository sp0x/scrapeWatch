package scrapeWatch

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	torrentdStatus "github.com/sp0x/torrentd/indexer/status"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func BindConfig() {
	viper.AutomaticEnv()
	_ = viper.BindEnv("firebase_credentials_file")
	_ = viper.BindEnv("verbose")
}

type FirebaseConfig struct {
	Project     string
	Credentials string
}

func GetFirebaseConfig() (*FirebaseConfig, error) {
	project := viper.GetString("GOOGLE_CLOUD_PROJECT")
	creds := viper.GetString("GOOGLE_APPLICATION_CREDENTIALS")
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
	verbose := viper.GetString("verbose")
	if verbose == "true" {
		log.SetLevel(log.DebugLevel)
	}
	fb, err := NewFirebaseFromEnv()
	if err != nil {
		fmt.Printf("error initializing firestore: %v\n", err)
		//os.Exit(1)
	}
	firebase = fb
}

func NonErrorStatusReceived(ctx context.Context, m PubSubMessage) error {
	initialize()
	message := torrentdStatus.ScrapeSchemeMessage{}
	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		return err
	}
	err = storeStatus(ctx, &message)
	if err != nil {
		log.Error(err)
	}
	return err
}

func storeStatus(ctx context.Context, message *torrentdStatus.ScrapeSchemeMessage) error {
	var err error
	if firebase == nil {
		return errors.New("firebase not initialized")
	}
	schemes := firebase.Collection("schemes")
	schemeKey := getSchemeKey(message)
	schemeDoc := schemes.Doc(schemeKey)
	_, err = schemeDoc.Create(ctx, message)

	if err != nil && status.Code(err) == codes.AlreadyExists {
		existing, err := schemeDoc.Get(ctx)
		if err != nil {
			return err
		}
		var existingScheme torrentdStatus.ScrapeSchemeMessage
		err = existing.DataTo(&existingScheme)
		if err != nil {
			return err
		}
		existingScheme.ResultsFound += message.ResultsFound
		existingScheme.Code = message.Code
		_, err = schemeDoc.Set(ctx, &existing)
		if err != nil {
			log.Debugf("Updated existing document %v.", schemeKey)
		}
		return err
	} else if err != nil {
		return err
	}
	log.Debugf("Created new document %v.", schemeKey)

	return err
}

func getSchemeKey(message *torrentdStatus.ScrapeSchemeMessage) string {
	if message.SchemeVersion != "" {
		return fmt.Sprintf("%s@%s", message.Site, message.SchemeVersion)
	} else {
		return message.Site
	}
}
