package main

import (

	//manager "./manager"
	"fmt"
	"log"

	manager "./manager/mongo"
	"github.com/ory/ladon"
	mgo "gopkg.in/mgo.v2"
)

func main() {

	session, err := mgo.Dial("127.0.0.1")
	db := session.DB("ladon").C("ladon")

	warden := &ladon.Ladon{
		Manager: manager.NewMongoManager(db),
	}

	/*
			var pol = &ladon.DefaultPolicy{
				// A required unique identifier. Used primarily for database retrieval.
				ID: "0000000-738b-41ec-b03c-b58a1b19d043",

				// A optional human readable description.
				Description: " ........ XYZ   >>>  .... something humanly readable",

				// A subject can be an user or a service. It is the "who" in "who is allowed to do what on something".
				// As you can see here, you can use regular expressions inside < >.
				Subjects: []string{"max", "peter", "<zac|ken>"},

				// Which resources this policy affects.
				// Again, you can put regular expressions in inside < >.
				Resources: []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},

				// Which actions this policy affects. Supports RegExp
				// Again, you can put regular expressions in inside < >.
				Actions: []string{"<create|delete>", "get"},

				// Should access be allowed or denied?
				// Note: If multiple policies match an access request, ladon.DenyAccess will always override ladon.AllowAccess
				// and thus deny access.
				Effect: ladon.AllowAccess,

				// Under which conditions this policy is "active".
				Conditions: ladon.Conditions{
					// In this example, the policy is only "active" when the requested subject is the owner of the resource as well.
					"resourceOwner": &ladon.EqualsSubjectCondition{},

					// Additionally, the policy will only match if the requests remote ip address matches address range 127.0.0.1/32
					"remoteIPAddress": &ladon.CIDRCondition{
						CIDR: "127.0.0.1/32",
					},
				},
			}

			err = warden.Manager.Update(pol)
			if err != nil {
				log.Println("ERROR 1:", err)
			}

				err = warden.Manager.Delete("0000000-738b-41ec-b03c-b58a1b19d043")
				if err != nil {
					log.Println("ERROR 2:", err)

				}

				all, err := warden.Manager.GetAll(10, 0)
				if err != nil {
					fmt.Println("ERR:", err)
				}
				log.Printf("%#v\n", all)



						p, err := warden.Manager.Get("68819e5a-738b-41ec-b03c-b58a1b19d043")
						if err != nil {
							log.Println("ERROR 2:", err)

						}


		p, err := warden.Manager.ColdStart()
		if err != nil {
			log.Println("ERROR 2:", err)

		}
		log.Printf("%#v\n", p)
	*/
	all, err := warden.Manager.GetAll(10, 0)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	log.Printf("%#v\n", all)

}
