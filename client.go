package pusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	AppId, Key, Secret string
}

func auth_timestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func (c *Client) trigger(channels []string, event string, _data interface{}, socket_id string) (error, string) {
	data, _ := json.Marshal(_data)

	payload, _ := json.Marshal(&Event{
		Name:     event,
		Channels: channels,
		Data:     string(data),
		SocketId: socket_id})

	path := "/apps/" + c.AppId + "/" + "events"

	u := CreateRequestUrl("POST", path, c.Key, c.Secret, auth_timestamp(), payload, nil)

	err, response := Request("POST", u, payload)

	return err, string(response)
}

func (c *Client) Trigger(channels []string, event string, _data interface{}) (error, string) {
	return c.trigger(channels, event, _data, "")
}

func (c *Client) TriggerExclusive(channels []string, event string, _data interface{}, socket_id string) (error, string) {
	return c.trigger(channels, event, _data, socket_id)
}

func (c *Client) Channels(additional_queries map[string]string) (error, *ChannelsList) {
	path := "/apps/" + c.AppId + "/channels"

	// fmt.Println("GET", path, c.Key, c.Secret, auth_timestamp(), nil, additional_queries)

	u := CreateRequestUrl("GET", path, c.Key, c.Secret, auth_timestamp(), nil, additional_queries)

	// fmt.Println(u)

	err, response := Request("GET", u, nil)

	channels := &ChannelsList{}
	json.Unmarshal(response, &channels)
	return err, channels
}

func (c *Client) Channel(name string, additional_queries map[string]string) (error, *Channel) {

	path := "/apps/" + c.AppId + "/channels/" + name

	// fmt.Println("GET", path, c.Key, c.Secret, auth_timestamp(), nil, additional_queries)

	u := CreateRequestUrl("GET", path, c.Key, c.Secret, auth_timestamp(), nil, additional_queries)

	// fmt.Println(u)

	err, raw_channel_data := Request("GET", u, nil)

	channel := &Channel{Name: name}
	json.Unmarshal(raw_channel_data, &channel)
	return err, channel

}

func (c *Client) GetChannelUsers(name string) (error, *Users) {
	path := "/apps/" + c.AppId + "/channels/" + name + "/users"

	fmt.Println("GET", path, c.Key, c.Secret, auth_timestamp(), nil, nil)

	u := CreateRequestUrl("GET", path, c.Key, c.Secret, auth_timestamp(), nil, nil)

	fmt.Println(u)

	err, raw_users := Request("GET", u, nil)
	users := &Users{}
	json.Unmarshal(raw_users, &users)
	return err, users
}

func (c *Client) AuthenticateChannel(_params []byte, presence_data MemberData) string {
	params, _ := url.ParseQuery(string(_params))
	channel_name := params["channel_name"][0]
	socket_id := params["socket_id"][0]

	string_to_sign := socket_id + ":" + channel_name

	is_presence_channel := strings.HasPrefix(channel_name, "presence-")

	var json_user_data string
	_response := make(map[string]string)

	if is_presence_channel {
		_json_user_data, _ := json.Marshal(presence_data)
		json_user_data = string(_json_user_data)
		string_to_sign += ":" + json_user_data

		_response["channel_data"] = json_user_data
	}

	auth_signature := HMACSignature(string_to_sign, c.Secret)
	_response["auth"] = c.Key + ":" + auth_signature
	response, _ := json.Marshal(_response)

	return string(response)
}

func (c *Client) Webhook(header http.Header, body []byte) *Webhook {
	webhook := &Webhook{Key: c.Key, Secret: c.Secret, Header: header, RawBody: string(body)}
	json.Unmarshal(body, &webhook)
	return webhook
}
