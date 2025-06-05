package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func infinitWaiting() {
	go func() {
		for {
			time.Sleep(time.Hour)
		}
	}()
	select {}
}

func main() {

	env := os.Getenv("UXP_TOKENS_PASS")
	if env == "" {
		fmt.Fprintln(os.Stderr, "UXP_TOKENS_PASS is not set")
		os.Exit(1)
	}

	// Разбиваем переменную окружения на пары
	pairs := strings.Split(env, ",")

	healthceck_status := false

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			fmt.Fprintf(os.Stderr, "Invalid token format: %s\n", pair)
			os.Exit(1)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Вызываем token_login -w
		fmt.Printf("Start login for token: %s\n", key)
		err := exec.Command("token_login", "-w", key, value).Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "token_login -w failed for %s: %v\n", key, err)
			infinitWaiting()
			//os.Exit(1)
		}

		// Вызываем token_login -r 10 раз
		fmt.Printf("Start login check for token: %s\n", key)
		for i := 0; i < 10; i++ {
			cmd_healthceck := exec.Command("trembita-healthcheck", "--log-level=fatal")
			cmd_healthceck.Stdout = os.Stdout
			cmd_healthceck.Stderr = os.Stderr
			if err := cmd_healthceck.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "trembita-healthcheck failed. Err: %v\n", err)
				healthceck_status = false
			} else {
				fmt.Fprintf(os.Stderr, "trembita-healthcheck succeeded!\n")
				healthceck_status = true
			}

			cmd := exec.Command("token_login", "-r", key)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "token_login -r failed at iteration %d for %s: %v\n", i+1, key, err)
				infinitWaiting()
				//os.Exit(1)
			}
			time.Sleep(1 * time.Second)
		}
	}

	if healthceck_status != true {
		fmt.Fprintf(os.Stderr, "trembita-healthcheck failed! Pod is not operatable! :( \n")
		infinitWaiting()
		//os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "trembita-healthcheck succeeded! :) \n")
	}

	fmt.Println("All tokens processed successfully.")
	os.Exit(0)
}
