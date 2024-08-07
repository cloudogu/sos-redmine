package redmine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type userResult struct {
	User User `json:"user"`
}

type usersResult struct {
	Users []User `json:"users"`
}

type User struct {
	Id           int            `json:"id"`
	Login        string         `json:"login"`
	Firstname    string         `json:"firstname"`
	Lastname     string         `json:"lastname"`
	Mail         string         `json:"mail"`
	CreatedOn    string         `json:"created_on"`
	LatLoginOn   string         `json:"last_login_on"`
	Memberships  []Membership   `json:"memberships"`
	CustomFields []*CustomField `json:"custom_fields,omitempty"`
}

type totalcount struct {
	TotalCount int `json:"total_count"`
}

type Status struct {
	User struct {
		Status int `json:"status"`
	} `json:"user"`
}

type UsersFilter struct {
	Filter
}

func NewUsersFilter() *UsersFilter {
	return &UsersFilter{Filter{}}
}

const (
	UserStatusAll        string = ""
	UserStatusActive     string = "1"
	UserStatusRegistered string = "2"
	UserStatusLocked     string = "3"
)

func (usf *UsersFilter) Status(status string) {
	usf.AddPair("status", status)
}

func (usf *UsersFilter) Name(name string) {
	usf.AddPair("name", name)
}

func (usf *UsersFilter) GroupId(groupId int) {
	usf.AddPair("group_id", strconv.Itoa(groupId))
}

type UserByIdFilter struct {
	Filter
}

func NewUserByIdFilter() *UserByIdFilter {
	return &UserByIdFilter{Filter{}}
}

const (
	UserIncludeMemberships string = "memberships"
	UserIncludeGroups      string = "groups"
)

func (uif *UserByIdFilter) Include(include string) {
	uif.AddPair("include", include)
}

func (c *Client) Users() ([]User, error) {
	res, err := c.Get(c.endpoint + "/users.json?key=" + c.apikey + c.getPaginationClause())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r usersResult
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return r.Users, nil
}

func (c *Client) totalCount() (int, error) {
	res, err := c.Get(c.endpoint + "/users.json?key=" + c.apikey + c.getPaginationClause())
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	var r totalcount
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)

	}
	if err != nil {
		return 0, err
	}
	return r.TotalCount, nil
}

func (c *Client) AllUsers() ([]User, error) {
	count, err := c.totalCount()
	if err != nil {
		return nil, err
	}
	var allUsers []User
	c.Limit = 100
	for i := 0; i < count; i = i + 100 {
		c.Offset = i
		users, err := c.Users()
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
	}
	c.Offset = 0
	return allUsers, nil
}

func (c *Client) SetUserStatus(status Status, userID int) error {
	s, err := json.Marshal(status)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(c.endpoint+"/users/%d.json?key="+c.apikey, userID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(s))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 204 {
		decoder := json.NewDecoder(res.Body)
		if !isHTTPStatusSuccessful(res.StatusCode, []int{http.StatusCreated}) {
			var er errorsResult
			err = decoder.Decode(&er)
			if err == nil {
				err = errors.New(strings.Join(er.Errors, "\n"))
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) UsersWithFilter(filter *UsersFilter) ([]User, error) {
	uri, err := c.URLWithFilter("/users.json", filter.Filter)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Redmine-API-Key", c.apikey)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r usersResult
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return r.Users, nil
}

func (c *Client) User(id int) (*User, error) {
	res, err := c.Get(c.endpoint + "/users/" + strconv.Itoa(id) + ".json?key=" + c.apikey)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r userResult
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return &r.User, nil
}

func (c *Client) UserByIdAndFilter(id int, filter *UserByIdFilter) (*User, error) {
	uri, err := c.URLWithFilter("/users/"+strconv.Itoa(id)+".json", filter.Filter)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Redmine-API-Key", c.apikey)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r userResult
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return &r.User, nil
}
