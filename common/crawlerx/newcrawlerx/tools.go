// Package newcrawlerx
// @Author bcy2007  2023/3/7 15:44
package newcrawlerx

import (
	"github.com/go-rod/rod"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"regexp"
	"strings"
	"time"
)

func getAttribute(element *rod.Element, attribute string) (string, error) {
	attributeStr, err := element.Attribute(attribute)
	if err != nil {
		return "", utils.Errorf("element %s get attribute error: %s", element, err)
	}
	if attributeStr == nil {
		return "", nil
	}
	return *attributeStr, nil
}

func getCurrentUrl(page *rod.Page) (string, error) {
	result, err := page.Eval(`()=>document.URL`)
	if err != nil {
		return "", utils.Errorf("page %s get url error: %s", page, err)
	}
	return result.Value.Str(), nil
}

func isVisible(element *rod.Element) (bool, error) {
	visible, err := element.Visible()
	if err != nil {
		return false, utils.Errorf("element %s get visiable error: %s", element, err)
	}
	return visible, nil
}

func getAllKeywords(element *rod.Element) string {
	var keywords string
	for _, attr := range elementAttribute {
		attribute, _ := getAttribute(element, attr)
		if attribute == "" {
			continue
		}
		keywords += attribute + ";"
	}
	return keywords
}

func getElementSelector(element *rod.Element) string {
	if visible, _ := element.Visible(); !visible {
		return ""
	}
	selectorObj, err := element.Eval(getSelector)
	if err != nil {
		log.Errorf("element %s get selector error: %s", element, err)
		return ""
	}
	return selectorObj.Value.Str()
}

func getElementsSelectors(elements rod.Elements) []string {
	selectors := make([]string, 0)
	for _, element := range elements {
		selector := getElementSelector(element)
		if selector == "" {
			continue
		}
		selectors = append(selectors, selector)
	}
	return selectors
}

func clickElementOnPageBySelector(page *rod.Page, selector string) {
	status, element, err := page.Has(selector)
	//element, err := page.Element(selector)
	if err != nil {
		log.Infof("on page element: %s", err)
		return
	}
	if !status {
		log.Infof("on page element: %s not found", selector)
	}
	if element == nil {
		log.Errorf("on page %s element %s not found.", page.MustInfo().URL, selector)
		return
	}
	if visible, _ := element.Visible(); !visible {
		return
	}
	//element.Click(proto.InputMouseButtonLeft)
	element.Eval(`this.click()`)
	page.MustWaitLoad()
	time.Sleep(500 * time.Millisecond)
}

func StringArrayContains(array []string, element string) bool {
	for _, s := range array {
		if element == s {
			return true
		}
	}
	return false
}

func isSimilarSelector(s1, s2 string) bool {
	if s1 == "" || s2 == "" {
		return false
	}
	sectionsA := strings.Split(s1, ">")
	sectionsB := strings.Split(s2, ">")
	if len(sectionsA) != len(sectionsB) {
		return false
	}
	flag := true
	length := len(sectionsA)
	for i := 0; i < length; i++ {
		if sectionsA[i] != sectionsB[i] {
			if !subCheck(sectionsA[i], sectionsB[i]) {
				return false
			}
			if flag == true {
				flag = false
			} else {
				return false
			}
		}
	}
	return true
}

func subCheck(s1, s2 string) bool {
	if s1 == "" || s2 == "" {
		return false
	}
	r, _ := regexp.Compile(`(\D+)?`)
	sectionA := r.FindAllString(s1, -1)
	sectionB := r.FindAllString(s2, -1)
	if len(sectionA) != len(sectionB) {
		return false
	}
	for i := 0; i < len(sectionA); i++ {
		if sectionA[i] != sectionB[i] {
			return false
		}
	}
	return true
}

func StringSuffixList(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

func StringPrefixList(origin string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(origin, prefix) {
			return true
		}
	}
	return false
}
