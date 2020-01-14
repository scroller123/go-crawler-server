package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	CONN_HOST         = ""
	CONN_PORT         = "9944"
	CONN_TYPE         = "tcp"
	NODE_WAIT_TIMEOUT = 30
)

type Response struct {
	Code    string
	Status  string
	Message string
	Uuid    string
}

type Command struct {
	Method     string
	Content    string
	Parameters string
	Async      bool
}

type ContextStorage struct {
	doc   *goquery.Document
	nodes map[string]*goquery.Selection
	mux   sync.Mutex
}

func Delimiter(base string) string {
	if len(base) > 0 {
		return "->"
	} else {
		return ""
	}
}

func (c *ContextStorage) SetVar(name string, nodeUuid string, i []int) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		c.mux.Lock()
		var istr = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i)), ";"), "[]")
		c.nodes[name+istr] = node
		c.mux.Unlock()

		return name, nil
	} else {
		return "", errors.New("goquery: SetVar nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Filter(filter string, nodeUuid string) (string, error) {
	key := nodeUuid + Delimiter(nodeUuid) + "filter(" + filter + ")"

	c.mux.Lock()
	node, ok := c.nodes[key]
	c.mux.Unlock()

	if ok { // нода уже есть в хранилище
		fmt.Println("Node ", key, " already done!")
		if node == nil {
			startAt := time.Now()
			for {
				fmt.Println("Waiting for node...")
				c.mux.Lock()
				node, _ = c.nodes[key]
				c.mux.Unlock()
				time.Sleep(time.Millisecond)
				if node != nil {
					break
				}
				if time.Since(startAt).Seconds() > NODE_WAIT_TIMEOUT {
					return strconv.Itoa(node.Length()), errors.New("goquery: Filter " + strconv.Itoa(NODE_WAIT_TIMEOUT) + " seconds wait left")
				}
			}
		}
		return strconv.Itoa(node.Length()), nil
	} else { // ноды нет в хранилище, создаем
		ts := time.Now()

		c.mux.Lock()
		node, ok = c.nodes[nodeUuid]
		c.mux.Unlock()

		if ok {
			var escaped, _ = url.QueryUnescape(filter)
			var find = node.Find(escaped)
			c.mux.Lock()
			c.nodes[key] = find
			c.mux.Unlock()
		} else {
			var escaped, _ = url.QueryUnescape(filter)
			var find = c.doc.Find(escaped)
			c.mux.Lock()
			c.nodes[key] = find
			c.mux.Unlock()
		}

		fmt.Printf("Parse filter '%s'  %f s [%s]\n", filter, time.Since(ts).Seconds(), nodeUuid)
		return strconv.Itoa(c.nodes[key].Length()), nil
	}
}

func (c *ContextStorage) Text(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if len(nodeUuid) == 0 {
		return c.doc.Text(), nil
	} else if ok {
		return node.Text(), nil
	} else {
		return "", errors.New("goquery: Text nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Html(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if len(nodeUuid) == 0 {
		html, _ := c.doc.Html()
		return html, nil
	} else if ok {
		html, _ := node.Html()
		return html, nil
	} else {
		return "", errors.New("goquery: Html nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Count(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		return strconv.Itoa(node.Length()), nil
	} else {
		return "", errors.New("goquery: Count nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Each(nodeUuid string, index int) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "each(" + strconv.Itoa(index) + ")"

		var find = node.Eq(index)

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: Each nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Attr(nodeUuid string, name string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		attr, _ := node.Attr(name)
		return attr, nil
	} else {
		return "", errors.New("goquery: Attr nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Eq(nodeUuid string, index int) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "eq(" + strconv.Itoa(index) + ")"

		var find = node.Eq(index)

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: Eq nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Last(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "last()"

		var find = node.Last()

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: Last nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) First(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "first()"

		var find = node.First()

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: First nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Siblings(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "siblings()"

		var find = node.Siblings()

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: Siblings nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) Children(nodeUuid string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "children()"

		var find = node.Children()

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: Children nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) GetNode(nodeUuid string, index int) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	c.mux.Unlock()

	if ok {
		key := nodeUuid + Delimiter(nodeUuid) + "getNode(" + strconv.Itoa(index) + ")"

		var find = node.Eq(index)

		c.mux.Lock()
		c.nodes[key] = find
		c.mux.Unlock()

		return key, nil
	} else {
		return "", errors.New("goquery: GetNode nodes not found via nodeUuid " + nodeUuid)
	}
}

func (c *ContextStorage) RemoveChild(nodeUuid string, nodeToRemove string) (string, error) {
	c.mux.Lock()
	node, ok := c.nodes[nodeUuid]
	nodeRem, okrem := c.nodes[nodeToRemove]
	c.mux.Unlock()

	if ok && okrem {
		key := nodeUuid + Delimiter(nodeUuid) + "removeChild(" + nodeToRemove + ")"

		nodeRem.Remove()

		c.mux.Lock()
		c.nodes[key] = node
		c.mux.Unlock()

		return key, nil
	} else {
		if !okrem {
			return "", errors.New("goquery: RemoveChild nodes not found via nodeToRemove " + nodeToRemove)
		} else {
			return "", errors.New("goquery: RemoveChild nodes not found via nodeUuid " + nodeUuid)
		}

	}
}

func main() {
	l, err := net.Listen(CONN_TYPE, ":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		fmt.Printf("Accepted connection %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		go handleRequest(conn)
	}
}

type Query struct {
	Type       string
	Query      string
	Parameters []string
}

type DomMap struct {
	Count string
	Html  string
	Text  string
}

type DomMapStorage struct {
	domMap map[string]DomMap
	varMap map[string]*goquery.Selection
	mux    sync.Mutex
}

func (storage *DomMapStorage) queryChain(base []string, chain []Query, context *ContextStorage, i []int) {
	for index, query := range chain {
		var prebase = append(base, query.Query+"("+strings.Join(query.Parameters, ";")+")")
		_, isSetvar := (storage.domMap)[strings.Join(base, "->")]
		if query.Type == "var" && isSetvar {
			context.SetVar(query.Query, strings.Join(base, "->"), i)
		}
		if query.Type == "method" {
			switch query.Query {
			case "filter":
				count, err := context.Filter(query.Parameters[0], strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				(storage.domMap)[strings.Join(prebase, "->")] = DomMap{
					Count: count,
					Html:  "",
					Text:  "",
				}

				// cnt, _ := strconv.Atoi(count)

				// for c := 0; c < cnt; c++ {
				// 	_, err = context.Eq(strings.Join(prebase, "->"), c)
				// 	if err != nil {
				// 		log.Println(err)
				// 		return
				// 	}
				// 	(storage.domMap)[strings.Join(prebase, "->")+"->eq("+strconv.Itoa(c)+")"] = DomMap{
				// 		Count: "",
				// 		Html:  "",
				// 		Text:  "",
				// 	}
				// 	var rebase = append(prebase, "eq("+strconv.Itoa(c)+")")
				// 	storage.queryChain(rebase, chain[index+1:], context)
				// }

			case "text":
				text, err := context.Text(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				// storage.mux.Lock()
				(storage.domMap)[strings.Join(base, "->")+"->text"] = DomMap{
					Count: "",
					Html:  "",
					Text:  text,
				}
				// storage.mux.Unlock()
			case "html":
				html, err := context.Html(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				// storage.mux.Lock()
				(storage.domMap)[strings.Join(base, "->")+"->html"] = DomMap{
					Count: "",
					Html:  html,
					Text:  "",
				}
				// storage.mux.Unlock()
			case "count":
				count, err := context.Count(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				// storage.mux.Lock()
				(storage.domMap)[strings.Join(base, "->")+"->count"] = DomMap{
					Count: count,
					Html:  "",
					Text:  "",
				}
				// storage.mux.Unlock()
			case "each":
				count, err := context.Count(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				// storage.mux.Lock()
				(storage.domMap)[strings.Join(base, "->")+"->each"] = DomMap{
					Count: count,
					Html:  "",
					Text:  "",
				}
				// storage.mux.Unlock()

				cnt, _ := strconv.Atoi(count)

				for c := 0; c < cnt; c++ {
					_, err = context.Each(strings.Join(base, "->"), c)
					if err != nil {
						log.Println(err)
						return
					}
					// storage.mux.Lock()
					(storage.domMap)[strings.Join(base, "->")+"->each("+strconv.Itoa(c)+")"] = DomMap{
						Count: "",
						Html:  "",
						Text:  "",
					}
					// storage.mux.Unlock()
					var rebase = append(base, query.Query+"("+strconv.Itoa(c)+")")
					storage.queryChain(rebase, chain[index+1:], context, append(i, c))
				}

				return
			case "eq":
				if len(query.Parameters) > 0 {
					var i, _ = strconv.Atoi(query.Parameters[0])
					var _, err = context.Eq(strings.Join(base, "->"), i)
					if err != nil {
						log.Println(err)
						break
					}
					// storage.mux.Lock()
					(storage.domMap)[strings.Join(base, "->")+"->eq("+strconv.Itoa(i)+")"] = DomMap{
						Count: "",
						Html:  "",
						Text:  "",
					}
					// storage.mux.Unlock()
				} else {
					count, err := context.Count(strings.Join(base, "->"))
					if err != nil {
						log.Println(err)
						return
					}
					// storage.mux.Lock()
					(storage.domMap)[strings.Join(base, "->")+"->eq"] = DomMap{
						Count: count,
						Html:  "",
						Text:  "",
					}
					// storage.mux.Unlock()
					cnt, _ := strconv.Atoi(count)

					for c := 0; c < cnt; c++ {
						_, err = context.Eq(strings.Join(base, "->"), c)
						if err != nil {
							log.Println(err)
							return
						}
						// storage.mux.Lock()
						(storage.domMap)[strings.Join(base, "->")+"->eq("+strconv.Itoa(c)+")"] = DomMap{
							Count: "",
							Html:  "",
							Text:  "",
						}
						// storage.mux.Unlock()
						var rebase = append(base, query.Query+"("+strconv.Itoa(c)+")")
						storage.queryChain(rebase, chain[index+1:], context, append(i, c))
					}

					return

				}
			case "attr":
				text, err := context.Attr(strings.Join(base, "->"), query.Parameters[0])
				if err != nil {
					log.Println(err)
					break
				}
				// storage.mux.Lock()
				(storage.domMap)[strings.Join(base, "->")+"->attr("+query.Parameters[0]+")"] = DomMap{
					Count: "",
					Html:  "",
					Text:  text,
				}
				// storage.mux.Unlock()
			case "last":
				_, err := context.Last(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				(storage.domMap)[strings.Join(base, "->")+"->last"] = DomMap{
					Count: "",
					Html:  "",
					Text:  "",
				}
			case "first":
				_, err := context.First(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				(storage.domMap)[strings.Join(base, "->")+"->first"] = DomMap{
					Count: "",
					Html:  "",
					Text:  "",
				}
			case "siblings":
				_, err := context.Siblings(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				(storage.domMap)[strings.Join(base, "->")+"->siblings"] = DomMap{
					Count: "",
					Html:  "",
					Text:  "",
				}
			case "children":
				_, err := context.Children(strings.Join(base, "->"))
				if err != nil {
					log.Println(err)
					return
				}
				(storage.domMap)[strings.Join(base, "->")+"->children"] = DomMap{
					Count: "",
					Html:  "",
					Text:  "",
				}
			case "getNode":
				if len(query.Parameters) == 0 {
					panic("getNode has no parameters [" + strings.Join(base, "->") + "]")
				}
				var i, _ = strconv.Atoi(query.Parameters[0])
				var key, err = context.GetNode(strings.Join(base, "->"), i)
				if err != nil {
					log.Println(err)
					return
				}

				var html string
				if len(query.Parameters) > 1 && query.Parameters[1] == "form" {
					var node = context.nodes[strings.Join(base, "->")]

					var action, _ = node.Attr("action")
					var id, _ = node.Attr("id")
					var class, _ = node.Attr("class")
					var method, _ = node.Attr("method")

					html = "<form method='" + method + "' class='" + class + "' id='" + id + "' action='" + action + "'>"

					var inputs = node.Find("input,select,button")

					for i, _ := range inputs.Nodes {
						single := inputs.Eq(i)
						v, _ := goquery.OuterHtml(single)
						html += v
					}
					html += "</form>"
				} else {
					html, _ = context.nodes[key].Html()
				}

				(storage.domMap)[strings.Join(base, "->")+"->getNode("+query.Parameters[0]+")"] = DomMap{
					Count: "",
					Html:  html,
					Text:  "",
				}
			case "removeChild":
				if len(query.Parameters) == 0 {
					panic("removeChild has no parameters [" + strings.Join(base, "->") + "]")
				}
				var istr = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i)), ";"), "[]")
				_, err := context.RemoveChild(strings.Join(base, "->"), query.Parameters[0]+istr)
				if err != nil {
					log.Println(err)
					return
				}

				(storage.domMap)[strings.Join(base, "->")+"->removeChild("+query.Parameters[0]+istr+")"] = DomMap{
					Count: "",
					Html:  "",
					Text:  "",
				}
			case "setAttribute":
			case "getAttribute":
			case "filterXpath":

			default:
			}
			base = append(base, query.Query+"("+strings.Join(query.Parameters, ";")+")")
		}
	}
}

func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}

func handleRequest(conn net.Conn) {

	defer conn.Close()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in handleRequest", r)
			fmt.Println(identifyPanic())
			pc, fn, line, _ := runtime.Caller(1)
			log.Printf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, r)
		}
	}()

	var context ContextStorage
	context.nodes = make(map[string]*goquery.Selection)

	var commandCounter int

	for {
		ts := time.Now()
		commandCounter++
		fmt.Println("Waiting for command... (", strconv.Itoa(commandCounter), ")")
		read, err := bufio.NewReader(conn).ReadString('\x06')
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF error")
			}
			fmt.Println("Read error")
			break
		}

		fmt.Printf("%f got message \n", time.Since(ts).Seconds())
		ts = time.Now()

		var response Response

		vars := strings.Split(strings.TrimRight(read, "\x06"), "\x05")
		async := false
		if vars[3] == "1" {
			async = true
		}
		command := Command{vars[0], vars[1], vars[2], async}
		fmt.Printf("Received command: %s, total bytes: %d bytes\n", command.Method, len(read))

		switch command.Method {
		case "addHtmlContent":
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(command.Content))
			if err != nil {
				log.Println(err)
				break
			}

			context.doc = doc

			var queries [][]Query

			err = json.Unmarshal([]byte(command.Parameters), &queries)
			if err != nil {
				fmt.Println("error:", err)
			}

			var storage DomMapStorage
			storage.domMap = make(map[string]DomMap)

			for _, chain := range queries {
				var base []string
				storage.queryChain(base, chain, &context, []int{})
			}

			var msg, _ = json.Marshal(storage.domMap)
			response.Message = string(msg)

			response.Status = "SUCCESS"
		// case "filter":
		// 	var nodeUuid string
		// 	if command.Async == true {
		// 		go context.Filter(command.Content, command.Parameters)
		// 		nodeUuid = ""
		// 	} else {
		// 		nodeUuid, err = context.Filter(command.Content, command.Parameters)
		// 	}

		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Uuid = nodeUuid
		// 	response.Status = "SUCCESS"
		// case "text":
		// 	text, err := context.Text(command.Content)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Status = "SUCCESS"
		// 	response.Message = text
		// case "html":
		// 	html, err := context.Html(command.Content)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Status = "SUCCESS"
		// 	response.Message = html
		// case "count":
		// 	html, err := context.Count(command.Content)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Status = "SUCCESS"
		// 	response.Message = html
		// case "each":
		// 	index, _ := strconv.Atoi(command.Parameters)
		// 	var nodeUuid string
		// 	nodeUuid, err := context.Each(command.Content, index)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Status = "SUCCESS"
		// 	response.Uuid = nodeUuid
		// case "attr":
		// 	text, err := context.Attr(command.Content, command.Parameters)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Status = "SUCCESS"
		// 	response.Message = text
		// case "eq":
		// 	index, _ := strconv.Atoi(command.Parameters)
		// 	var nodeUuid string
		// 	nodeUuid, err := context.Eq(command.Content, index)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		break
		// 	}
		// 	response.Status = "SUCCESS"
		// 	response.Uuid = nodeUuid
		default:
			response.Status = "ERROR"
			response.Message = "Method not found"
		}

		fmt.Println("Response: ", response.Code, ", ", response.Status, ", ", response.Uuid)
		conn.Write([]byte(response.Code + "\x05"))
		conn.Write([]byte(response.Status + "\x05"))
		conn.Write([]byte(response.Message + "\x05"))
		conn.Write([]byte(response.Uuid))
		conn.Write([]byte("\x06"))

		fmt.Printf("%f total  \n", time.Since(ts).Seconds())

	}
}
