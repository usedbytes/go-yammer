package yammer

import (
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bmorton/go-yammer/schema"
)

type ThreadingMode string
const (
	ThreadExtended = "extended"
	ThreadTopOnly =  "true"
)

type FeedParams struct {
	Older_than, Newer_than int
	Threaded ThreadingMode
	Limit int
}

func (p *FeedParams) Values() url.Values {
	v := url.Values{}
	if p.Older_than != 0 {
		v.Add("older_than", strconv.Itoa(p.Older_than))
	}
	if p.Newer_than != 0 {
		v.Add("newer_than", strconv.Itoa(p.Newer_than))
	}
	if p.Threaded != "" {
		v.Add("threaded", string(p.Threaded))
	}
	if p.Limit != 0 {
		v.Add("limit", strconv.Itoa(p.Limit))
	}
	return v
}

func (c *Client) GroupFeed(id int) (*schema.MessageFeed, error) {
	var parms FeedParams

	return c.GroupFeedParams(id, parms)
}

func (c *Client) GroupFeedParams(id int, parms FeedParams) (*schema.MessageFeed, error) {
	url := fmt.Sprintf("%s/api/v1/messages/in_group/%d.json", c.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &schema.MessageFeed{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))

	args := req.URL.Query()
	extra_args := parms.Values()
	for k, vals := range extra_args {
		for _, v := range vals {
			args.Add(k, v)
		}
	}
	req.URL.RawQuery = args.Encode()

	resp, err := c.connection.Do(req)
	if err != nil {
		log.Println(err)
		return &schema.MessageFeed{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return &schema.MessageFeed{}, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &schema.MessageFeed{}, err
	}

	var feed schema.MessageFeed
	err = json.Unmarshal(body, &feed)
	if err != nil {
		return &schema.MessageFeed{}, err
	}

	return &feed, nil
}

func (c *Client) GroupFeedSince(id int, newerThan int) (*schema.MessageFeed, error) {
	var combined *schema.MessageFeed

	if newerThan == 0 {
		return &schema.MessageFeed{}, nil
	}

	params := FeedParams{
		Newer_than: newerThan,
	}

	for {
		feed, err := c.GroupFeedParams(id, params)
		if err != nil {
			return combined, err
		}
		if len(feed.Messages) == 0 {
			return combined, err
		}

		if combined != nil {
			combined.Messages = append(combined.Messages, feed.Messages...)
			combined.References = append(combined.References, feed.References...)
			/* TODO: Meta/ */
		} else {
			combined = feed
		}

		params.Older_than = feed.Messages[len(feed.Messages)-1].Id
	}
}


func (c *Client) InboxFeed() (*schema.MessageFeed, error) {
	url := fmt.Sprintf("%s/api/v1/messages/inbox.json", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &schema.MessageFeed{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))

	resp, err := c.connection.Do(req)
	if err != nil {
		log.Println(err)
		return &schema.MessageFeed{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return &schema.MessageFeed{}, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &schema.MessageFeed{}, err
	}

	var feed schema.MessageFeed
	err = json.Unmarshal(body, &feed)
	if err != nil {
		return &schema.MessageFeed{}, err
	}

	return &feed, nil
}
