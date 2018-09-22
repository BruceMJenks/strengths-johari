package main

import "fmt"

// WindowPanes represents the johari window panes
type WindowPanes struct {
	Arena   []string `json:"arena"`
	Blind   []string `json:"blind"`
	Facade  []string `json:"facade"`
	Unknown []string `json:"unknown"`
}

// JCWindows marries joahir and clifton panes
type JCWindows struct {
	Johari  WindowPanes `json:"johari"`
	Clifton WindowPanes `json:"clifton"`
}

// GetWindowPanesFromDB generate JCWindows struct from databse
func GetWindowPanesFromDB(res *JCWindows, uid int, sess string) error {
	jaq := fmt.Sprintf(JOHARI_ARENA_QUERY, sess)
	jbq := fmt.Sprintf(JOHARI_BLIND_QUERY, sess, sess)
	jfq := fmt.Sprintf(JOHARI_FACADE_QUERY, sess, sess)
	juq := fmt.Sprintf(JOHARI_UNKOWN_QUERY, sess, sess)
	caq := fmt.Sprintf(CLIFTON_ARENA_QUERY, sess)
	cbq := fmt.Sprintf(CLIFTON_BLIND_QUERY, sess, sess)
	cfq := fmt.Sprintf(CLIFTON_FACADE_QUERY, sess, sess)
	cuq := fmt.Sprintf(CLIFTON_UNKOWN_QUERY, sess, sess)
	var err error

	res.Johari.Arena, err = DBI.GetStringList(jaq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", jaq, err)
		return err
	}
	res.Johari.Blind, err = DBI.GetStringList(jbq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", jbq, err)
		return err
	}
	res.Johari.Facade, err = DBI.GetStringList(jfq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", jfq, err)
		return err
	}
	res.Johari.Unknown, err = DBI.GetStringList(juq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", juq, err)
		return err
	}

	res.Clifton.Arena, err = DBI.GetStringList(caq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", caq, err)
		return err
	}
	res.Clifton.Blind, err = DBI.GetStringList(cbq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", cbq, err)
		return err
	}
	res.Clifton.Facade, err = DBI.GetStringList(cfq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", cfq, err)
		return err
	}
	res.Clifton.Unknown, err = DBI.GetStringList(cuq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", cuq, err)
		return err
	}

	//Normalize the themes and make sure there are not duplicates in CLIFTON_UNKOWN_QUERY and CLIFTON_FACADE_QUERY
	knownWords := make([]string, 0)
	knownWords = append(knownWords, res.Clifton.Arena...)
	knownWords = append(knownWords, res.Clifton.Blind...)
	res.Clifton.Unknown = pruneCliftonUnkown(res.Clifton.Unknown, knownWords)
	res.Clifton.Facade = pruneCliftonUnkown(res.Clifton.Facade, knownWords)

	return nil
}

func pruneCliftonUnkown(u, k []string) []string {
	n := make([]string, 0)
	matches := make(map[string]bool)
	for i := range u {
		matches[u[i]] = false
	}

	for i := range u {
		for x := range k {
			if u[i] == k[x] {
				matches[u[i]] = true
			}
		}
	}

	for k, v := range matches {
		if !v {
			n = append(n, k)
		}
	}
	return n
}
