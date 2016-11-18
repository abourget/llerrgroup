package llerrgroup

import (
	"log"
	"net/http"
)

func ExampleGroupParallelAndQuitSoon() {
	url := []string{"http://1", "http://2"}
	grp := New(15)
	for _, u := range url {
		if grp.Stop() {
			continue
		}

		u := u
		grp.Go(func() error {
			log.Println("Fetching", u)
			_, err := http.Get(u)
			return err
		})
	}
	err := grp.Wait()
	if err != nil {
		log.Println("Errored:", err)
	}
	log.Println("All done!")
}
