package service

import "strings"

func mapFromSlice(vals []string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, v := range vals {
		res[v] = struct{}{}
	}
	return res
}

func getOwnersFromAnnotations(anns annotations) []string {
	if anns == nil {
		return nil
	}
	ownersStr := strings.Replace(anns[ownersKey], " ", "", -1)
	if len(ownersStr) == 0 {
		return nil
	}
	return strings.Split(ownersStr, ",")
}

func isOwner(user string, owners []string) bool {
	for _, v := range owners {
		if v == user {
			return true
		}
	}
	return false
}

func removeRepeatedItems(items []string) []string {
	res := make([]string, 0, len(items))
	keys := make(map[string]struct{})
	flag := struct{}{}
	var rep bool
	for _, v := range items {
		_, rep = keys[v]
		if rep {
			continue
		}
		res = append(res, v)
		keys[v] = flag
	}
	return res
}

func removeItemFromItems(item string, items []string) []string {
	res := make([]string, 0, len(items))
	for _, v := range items {
		if v == item {
			continue
		}
		res = append(res, v)
	}
	return res
}
