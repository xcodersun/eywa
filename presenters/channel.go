package presenters

import (
	"text/template"
	"bytes"
	"path"
	"fmt"
	"strings"
	"github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/utils"
	. "github.com/vivowares/eywa/loggers"
	. "github.com/vivowares/eywa/models"
)

type HardwareTemplateHandler func(ch *Channel)(string, error)

type ChannelBrief struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChannelDetail struct {
	ID string `json:"id"`
	*Channel
}

var SupportedHardwareTemplateLanguages = []string{"clang"}
var hardwareTemplateExtensions = map[string]string {
	"clang": "%s.h",
}
var hardwareTemplateHandlers = map[string]HardwareTemplateHandler {
    "clang": fetchClangTemplateContentByChannel,
}

func NewChannelBrief(c *Channel) *ChannelBrief {
	hashId, _ := c.HashId()
	return &ChannelBrief{
		ID:          hashId,
		Name:        c.Name,
		Description: c.Description,
	}
}

func NewChannelDetail(c *Channel) *ChannelDetail {
	hashId, _ := c.HashId()
	return &ChannelDetail{
		ID:      hashId,
		Channel: c,
	}
}

func FetchHardwareTemplateContentByChannel(ch *Channel, lang string) (string, string, error) {
	fileName := fmt.Sprintf(hardwareTemplateExtensions[lang], ch.Name)
	fileName = strings.Replace(fileName, " ", "_", -1)

	content, err := hardwareTemplateHandlers[lang](ch)

	return fileName, content, err
}

func fetchClangTemplateContentByChannel(ch *Channel) (string, error) {
	var bufHeader, bufBody bytes.Buffer

	hwTmplPath := path.Join(configs.Config().Service.Assets, "hardware_templates",
							fmt.Sprintf(hardwareTemplateExtensions["clang"], "clang"))

	header, err := utils.HardwareTemplateParse(hwTmplPath, "CLANG_HTTP_POST_HEADER", "#defkey", "#end")
	if err != nil {
		Logger.Error(fmt.Sprintf("%v", err))
		return "", err
	}

	tmplHeader, err := template.New("clang_http_post_template").Parse(header)
	if err != nil {
		Logger.Error(fmt.Sprintf("%v", err))
		return "", err
	}
	err = tmplHeader.Execute(&bufHeader, ch)
	if err != nil {
		Logger.Error(fmt.Sprintf("%v", err))
		return "", err
	}

	body, err := utils.HardwareTemplateParse(hwTmplPath, "CLANG_HTTP_POST_BODY", "#defkey", "#end")
	if err != nil {
		Logger.Error(fmt.Sprintf("%v", err))
		return "", err
	}

	tmplBody, err := template.New("clang_http_post_body").Parse(body)
	if err != nil {
		Logger.Error(fmt.Sprintf("%v", err))
		return "", err
	}
	err = tmplBody.Execute(&bufBody, ch)
	if err != nil {
		Logger.Error(fmt.Sprintf("%v", err))
		return "", err
	}

	content := bufHeader.String() + strings.Replace(bufBody.String(), ",}\\r\\n", "}\\r\\n", 1)

	return content, err
}
