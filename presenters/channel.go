package presenters

import (
	"text/template"
	"bytes"
	"path"
	"strings"
	"github.com/eywa/configs"
	"github.com/eywa/utils"
	. "github.com/eywa/models"
)

type HardwareTemplateHandler func(ch *Channel)(string, error)

type ChannelBrief struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Created     int64  `json:"created"`
}

type ChannelDetail struct {
	ID string `json:"id"`
	*Channel
}

func NewChannelBrief(c *Channel) *ChannelBrief {
	hashId, _ := c.HashId()
	return &ChannelBrief{
		ID:          hashId,
		Name:        c.Name,
		Description: c.Description,
		Created:     c.Created,
	}
}

func NewChannelDetail(c *Channel) *ChannelDetail {
	hashId, _ := c.HashId()
	return &ChannelDetail{
		ID:      hashId,
		Channel: c,
	}
}

func FetchRequestTemplateByChannel(ch *Channel) (f string, t string, e error) {
	var bufHeader, bufBody bytes.Buffer

	tmplName := strings.Replace(ch.Name, " ", "_", -1)

	hwTmplPath := path.Join(configs.Config().Service.Templates, "request.tmpl")

	header, err := utils.RequestTemplateParse(hwTmplPath, "HTTP_POST_HEADER", "#defkey", "#end")
	if err != nil {
		return "", "", err
	}

	tmplHeader, err := template.New("http_post_header").Parse(header)
	if err != nil {
		return "", "", err
	}
	err = tmplHeader.Execute(&bufHeader, ch)
	if err != nil {
		return "", "", err
	}

	body, err := utils.RequestTemplateParse(hwTmplPath, "HTTP_POST_BODY", "#defkey", "#end")
	if err != nil {
		return "", "", err
	}

	tmplBody, err := template.New("http_post_body").Parse(body)
	if err != nil {
		return "", "", err
	}
	err = tmplBody.Execute(&bufBody, ch)
	if err != nil {
		return "", "", err
	}

	tmpl := bufHeader.String() + strings.Replace(bufBody.String(), ",}", "}", 1)

	return tmplName, tmpl, err
}
