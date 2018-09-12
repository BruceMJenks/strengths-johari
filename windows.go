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

	res.Johari.Arena, err = dbi.GetStringList(jaq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", jaq, err)
		return err
	}
	res.Johari.Blind, err = dbi.GetStringList(jbq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", jbq, err)
		return err
	}
	res.Johari.Facade, err = dbi.GetStringList(jfq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", jfq, err)
		return err
	}
	res.Johari.Unknown, err = dbi.GetStringList(juq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", juq, err)
		return err
	}

	res.Clifton.Arena, err = dbi.GetStringList(caq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", caq, err)
		return err
	}
	res.Clifton.Blind, err = dbi.GetStringList(cbq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", cbq, err)
		return err
	}
	res.Clifton.Facade, err = dbi.GetStringList(cfq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", cfq, err)
		return err
	}
	res.Clifton.Unknown, err = dbi.GetStringList(cuq)
	if err != nil {
		fmt.Printf("Query:%s\nError:%s\n", cuq, err)
		return err
	}
	return nil
}
