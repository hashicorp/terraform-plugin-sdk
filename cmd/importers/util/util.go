package util

import (
	"log"
	"os"
)

func MustEnv(env string) string {
	v := os.Getenv(env)
	if v == "" {
		log.Fatalf("Env Var %q must be set", env)
	}
	return v
}

func StringListToSet(list []string) map[string]bool {
	res := make(map[string]bool)
	for _, k := range list {
		res[k] = true
	}
	return res
}

func StringSetToList(set map[string]bool) []string {
	res := make([]string, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	return res
}
