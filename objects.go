package main


type sectask struct {
	nic_id string
	ip_add map[string]bool
}
func (task *sectask) add_ips_to_nic() {
}

type delip struct {
	nic_id string
	ip_rem map[string]bool
}
func (task *sectask) rem_ips_from_nic() {
	// Delete data nic if the secip count - 0
}

type sgtask struct {
	sgid string
	rules []sgrule
}
type sgrule struct {
	sport int64
	eport int64
	proto string
	cidr string
}

// Release fip to be called in delete_vip
