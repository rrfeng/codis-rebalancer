package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var d *string = flag.String("d", "127.0.0.1:18080", "Dashboard Addr.")
var i *int = flag.Int("i", 10, "Migrate interval.")
var f *bool = flag.Bool("f", false, "Do the action. Default only show the actions.")

type slotInfo struct {
	Id       int         `json:"id"`
	Group_id int         `json:"group_id"`
	Action   interface{} `json:"-"`
}

type groupInfo struct {
	Id int `json:"id"`
}

type groupModel struct {
	Models []groupInfo `json:"models"`
}

type stats struct {
	Closed      bool       `json:"closed"`
	Slots       []slotInfo `json:"slots"`
	Group       groupModel `json:"group"`
	proxy       interface{}
	slot_action interface{}
}

type config struct {
	Coordinator_name string `json:"coordinator_name"`
	Coordinator_addr string `json:"coordinator_addr"`
	Admin_addr       string `json:"admin_addr"`
	Product_name     string `json:"product_name"`
}

type codis struct {
	version string
	compile string
	Config  config `json:"config"`
	model   interface{}
	Stats   stats `json:"stats"`
}

func main() {
	flag.Parse()
	if *d == "" {
		log.Fatalln("Please provide the dashboard addr!")
	}

	url := "http://" + *d + "/topom"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln("Get stats from dashboard error: ", err.Error())
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatalln("Read response error: ", err.Error())
	}
	var s codis
	err = json.Unmarshal(b, &s)
	if err != nil {
		log.Fatalln("Json decode response error: ", err.Error())
	}

	// from groups get all group, from slots get which slots in the group.
	var groups = map[int][]int{}
	for _, g := range s.Stats.Group.Models {
		groups[g.Id] = []int{}
	}
	for _, slot := range s.Stats.Slots {
		groups[slot.Group_id] = append(groups[slot.Group_id], slot.Id)
	}

	// calc which slots to migrate, which groups to migrate to.
	bala := balancer(len(groups))
	slotPool := []int{}
	targetGroup := map[int]int{}
	i := 0
	for gid, slots := range groups {
		if len(slots) > bala[i] {
			to_remove := pickSlots(slots, bala[i])
			slotPool = append(slotPool, to_remove...)
		} else if len(slots) < bala[i] {
			targetGroup[gid] = bala[i] - len(slots)
		}
		i++
	}

	// do the migrate action.
	xauth := genAuth(s.Config.Product_name)
	client := &http.Client{}
	err = setInterval(i, client, xauth, *d)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for gid, delta := range targetGroup {
		for i := 0; i < delta; i++ {
			if *f == true {
				err := migrate(slotPool[0], gid, client, xauth, *d)
				if err != nil {
					log.Fatalf("Error migrating slot %d to group %d: %s\n", slotPool[0], gid, err.Error())
				} else {
					slotPool = slotPool[1:]
					log.Printf("migrate one slot to group %d success, %d to migrate.", gid, len(slotPool))
				}
			} else {
				slotPool = slotPool[1:]
				log.Printf("migrate one slot to group %d success, %d to migrate.", gid, len(slotPool))
			}
		}
	}
}

func balancer(n int) []int {
	var base, remi int
	var a = make([]int, n)
	base = 1024 / n
	remi = 1024 % n
	for i := 0; i < n; i++ {
		if i < remi {
			a[i] = base + 1
		} else {
			a[i] = base
		}
	}
	return a
}

func pickSlots(g []int, tn int) []int {
	if tn > len(g) {
		return []int{}
	} else {
		return g[tn:]
	}
}

func genAuth(name string) string {
	s := []byte("Codis-XAuth-[" + name + "]")
	md := sha256.Sum256(s)
	mdstr := hex.EncodeToString(md[:32])
	return mdstr[:32]
}

func setInterval(interval int, client *http.Client, xauth string, addr string) error {
	url := "http://" + addr + "/api/topom/slots/action/interval/" + xauth + "/" + strconv.Itoa(interval)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("response status: %d, response body: %s", resp.StatusCode, string(b)))
	}
	resp.Body.Close()
	return nil
}

func migrate(sid, gid int, client *http.Client, xauth string, addr string) error {
	url := "http://" + addr + "/api/topom/slots/action/create/" + xauth + "/" + strconv.Itoa(sid) + "/" + strconv.Itoa(gid)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("response status: %d, response body: %s", resp.StatusCode, string(b)))
	}
	resp.Body.Close()
	return nil
}
