package main

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// AdminInput is the complete, public, declarative material an administrator
// supplies to instantiate this implementation. It carries no private key, seed,
// staker file, signer file or API credential, and it never carries core
// references: those are injected from the compiled CORE_BINDING.
type AdminInput struct {
	Schema        string             `json:"schema"`
	Version       uint32             `json:"version"`
	Instantiation AdminInstantiation `json:"instantiation"`
	VM            AdminVM            `json:"vm"`
	Architecture  AdminArchitecture  `json:"architecture"`
	Control       AdminControl       `json:"control"`
	Authority     AdminAuthority     `json:"authority"`
	Validators    []AdminValidator   `json:"validators"`
	Declarations  AdminDeclarations  `json:"declarations"`
}

type AdminInstantiation struct {
	FinalName  string `json:"final_name"`
	LocalityID string `json:"locality_id"`
	GenesisID  string `json:"genesis_id"`
	Purpose    string `json:"purpose"`
	Network    string `json:"network"`
}

type AdminVM struct {
	Name    string `json:"vm_name"`
	ID      string `json:"vm_id"`
	Version string `json:"vm_version"`
}

type AdminArchitecture struct {
	ManagerArchitecture  string `json:"manager_architecture"`
	ManagerChain         string `json:"validator_manager_chain"`
	ManagerAddress       string `json:"validator_manager_address"`
	ManagerOwnerAddress  string `json:"validator_manager_owner_address"`
	HostPlatform         string `json:"host_platform"`
	PluginDirectory      string `json:"validator_plugin_directory"`
	AvalancheGoVersion   string `json:"avalanchego_version"`
	RPCChainVMProtocolID uint32 `json:"rpcchainvm_protocol"`
}

type AdminControl struct {
	PChainControllerAddress string     `json:"p_chain_controller_address"`
	SubnetOwner             AdminOwner `json:"subnet_owner"`
	RemainingBalanceOwner   AdminOwner `json:"remaining_balance_owner"`
	DisableOwner            AdminOwner `json:"disable_owner"`
}

type AdminOwner struct {
	Addresses []string `json:"addresses"`
	Threshold uint32   `json:"threshold"`
}

type AdminAuthority struct {
	FormationPublicKey   string   `json:"formation_public_key"`
	ExhaustionPlanned    bool     `json:"formation_authority_exhaustion_planned"`
	ContinuityPublicKeys []string `json:"continuity_public_keys"`
}

type AdminValidator struct {
	NodeID                 string `json:"node_id"`
	BLSPublicKey           string `json:"bls_public_key"`
	BLSProofOfPossession   string `json:"bls_proof_of_possession"`
	Weight                 uint64 `json:"weight"`
	InitialBalanceNanoAVAX uint64 `json:"initial_balance_nanoavax"`
	RunwayDays             uint32 `json:"runway_days"`
	PublicP2PEndpoint      string `json:"public_p2p_endpoint"`
	HostPlatform           string `json:"host_platform"`
	KeyCustody             string `json:"key_custody"`
}

type AdminDeclarations struct {
	ValidatorKeyCustody   string `json:"validator_key_custody"`
	BalanceFunding        string `json:"balance_funding"`
	ConversionAwareness   string `json:"conversion_irreversibility_acknowledged"`
	ManagerInitialization string `json:"manager_initialization_is_one_shot_acknowledged"`
}

const (
	networkMainnet   = "MAINNET"
	networkLocalTest = "LOCAL_TEST"

	managerArchitecture = "C_CHAIN_EVM"
	managerChain        = "C_CHAIN"
	supportedPlatform   = "linux-amd64"
)

var (
	nodeIDPattern       = regexp.MustCompile(`^NodeID-[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]{30,50}$`)
	evmAddressPattern   = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	pChainAddressRegexp = regexp.MustCompile(`^P-[a-z0-9]{10,90}$`)
	endpointPattern     = regexp.MustCompile(`^[A-Za-z0-9.\-]{3,253}:[0-9]{2,5}$`)
	hexPattern          = regexp.MustCompile(`^[0-9a-f]+$`)
	prefixedHexPattern  = regexp.MustCompile(`^0x[0-9a-f]+$`)

	// placeholderMarkers are refused anywhere in administrator input. Instantiation
	// must not proceed on unfinished material.
	placeholderMarkers = []string{
		"PLACEHOLDER", "TODO", "TBD", "CHANGEME", "CHANGE_ME", "FIXME", "EXAMPLE",
		"SAMPLE", "REPLACE_ME", "REPLACEME", "XXXX", "YOUR_", "DUMMY", "FAKE", "TEST_VALUE",
	}

	// forbiddenKeyMarkers are refused as administrator input keys. Private custody
	// material never enters this implementation.
	forbiddenKeyMarkers = []string{
		"private", "secret", "seed", "mnemonic", "passphrase", "password",
		"staker_key", "staker-key", "signer_key", "signer-key", "signing_key",
		"api_key", "apikey", "api_token", "access_token", "credential", "keystore",
		"wallet_file", "ewoq", "bearer",
	}
)

func parseAdminInput(raw []byte) (*AdminInput, error) {
	if err := refuseForbiddenKeys(raw); err != nil {
		return nil, err
	}
	input := new(AdminInput)
	if err := decodeStrict(raw, input); err != nil {
		return nil, fmt.Errorf("decode administrator input: %w", err)
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	return input, nil
}

// refuseForbiddenKeys walks the raw document and refuses any key that could
// carry private custody material, before any value is interpreted.
func refuseForbiddenKeys(raw []byte) error {
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("administrator input is not valid JSON: %w", err)
	}
	return walkKeys(document, func(key string) error {
		lowered := strings.ToLower(key)
		for _, marker := range forbiddenKeyMarkers {
			if strings.Contains(lowered, marker) {
				return fmt.Errorf("administrator input must never carry %q material", key)
			}
		}
		return nil
	})
}

func walkKeys(node any, check func(string) error) error {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			if err := check(key); err != nil {
				return err
			}
			if err := walkKeys(value, check); err != nil {
				return err
			}
		}
	case []any:
		for _, value := range typed {
			if err := walkKeys(value, check); err != nil {
				return err
			}
		}
	}
	return nil
}

func refusePlaceholder(field, value string) error {
	upper := strings.ToUpper(value)
	for _, marker := range placeholderMarkers {
		if strings.Contains(upper, marker) {
			return fmt.Errorf("%s still carries the unfinished marker %s", field, marker)
		}
	}
	return nil
}

func refuseZeroDigest(field, value string) error {
	trimmed := strings.TrimPrefix(strings.ToLower(value), "0x")
	if trimmed == "" {
		return fmt.Errorf("%s is empty", field)
	}
	if strings.Trim(trimmed, "0") == "" {
		return fmt.Errorf("%s is an all-zero value", field)
	}
	return nil
}

func requirePrefixedHex(field, value string, byteLength int) error {
	if err := refusePlaceholder(field, value); err != nil {
		return err
	}
	if !prefixedHexPattern.MatchString(value) || len(value) != 2+byteLength*2 {
		return fmt.Errorf("%s must be 0x-prefixed lowercase hexadecimal of %d bytes", field, byteLength)
	}
	return refuseZeroDigest(field, value)
}

func requirePlainHex(field, value string, byteLength int) error {
	if err := refusePlaceholder(field, value); err != nil {
		return err
	}
	if !hexPattern.MatchString(value) || len(value) != byteLength*2 {
		return fmt.Errorf("%s must be lowercase hexadecimal of %d bytes", field, byteLength)
	}
	return refuseZeroDigest(field, value)
}

func requireEVMAddress(field, value string) error {
	if err := refusePlaceholder(field, value); err != nil {
		return err
	}
	if !evmAddressPattern.MatchString(value) {
		return fmt.Errorf("%s must be a 0x-prefixed 20-byte address", field)
	}
	return refuseZeroDigest(field, value)
}

func (a *AdminInput) Validate() error {
	if a.Schema != adminInputSchema || a.Version != 1 {
		return fmt.Errorf("expected administrator input schema %s version 1", adminInputSchema)
	}
	if err := a.Instantiation.validate(); err != nil {
		return err
	}
	if err := a.VM.validate(); err != nil {
		return err
	}
	if err := a.Architecture.validate(); err != nil {
		return err
	}
	if err := a.Control.validate(); err != nil {
		return err
	}
	if err := a.Authority.validate(); err != nil {
		return err
	}
	if err := a.validateValidators(); err != nil {
		return err
	}
	return a.Declarations.validate()
}

func (i *AdminInstantiation) validate() error {
	if err := requireBoundedText("instantiation.final_name", i.FinalName, 128); err != nil {
		return err
	}
	if err := refusePlaceholder("instantiation.final_name", i.FinalName); err != nil {
		return err
	}
	for field, value := range map[string]string{
		"instantiation.locality_id": i.LocalityID,
		"instantiation.genesis_id":  i.GenesisID,
	} {
		if err := requireSafeID(field, value); err != nil {
			return err
		}
		if err := refusePlaceholder(field, value); err != nil {
			return err
		}
		if err := requireDistinctFromPredecessors(field, value); err != nil {
			return err
		}
	}
	if i.LocalityID == i.GenesisID {
		return errors.New("instantiation.locality_id and instantiation.genesis_id must differ")
	}
	if err := requireBoundedText("instantiation.purpose", i.Purpose, 1024); err != nil {
		return err
	}
	if err := refusePlaceholder("instantiation.purpose", i.Purpose); err != nil {
		return err
	}
	if i.Network != networkMainnet && i.Network != networkLocalTest {
		return fmt.Errorf("instantiation.network must be %s or %s", networkMainnet, networkLocalTest)
	}
	return nil
}

func (v *AdminVM) validate() error {
	if v.Name != vmName {
		return fmt.Errorf("vm.vm_name must be %s", vmName)
	}
	if v.ID != expectedVMID {
		return fmt.Errorf("vm.vm_id must be the recorded identifier %s", expectedVMID)
	}
	if v.Version != vmVersion {
		return fmt.Errorf("vm.vm_version must be %s", vmVersion)
	}
	return nil
}

func (a *AdminArchitecture) validate() error {
	if a.ManagerArchitecture != managerArchitecture {
		return fmt.Errorf("architecture.manager_architecture is unresolved: only %s is supported by this implementation", managerArchitecture)
	}
	if a.ManagerChain != managerChain {
		return fmt.Errorf("architecture.validator_manager_chain must be %s", managerChain)
	}
	if err := requireEVMAddress("architecture.validator_manager_address", a.ManagerAddress); err != nil {
		return err
	}
	if err := requireEVMAddress("architecture.validator_manager_owner_address", a.ManagerOwnerAddress); err != nil {
		return err
	}
	if strings.EqualFold(a.ManagerAddress, a.ManagerOwnerAddress) {
		return errors.New("architecture.validator_manager_owner_address must not be the manager contract address")
	}
	if a.HostPlatform != supportedPlatform {
		return fmt.Errorf("architecture.host_platform %q is unsupported: this implementation is qualified for %s only", a.HostPlatform, supportedPlatform)
	}
	if !strings.HasPrefix(a.PluginDirectory, "/") || strings.Contains(a.PluginDirectory, "..") {
		return errors.New("architecture.validator_plugin_directory must be an absolute path without parent references")
	}
	if err := refusePlaceholder("architecture.validator_plugin_directory", a.PluginDirectory); err != nil {
		return err
	}
	if a.AvalancheGoVersion != avalancheGoVersion {
		return fmt.Errorf("architecture.avalanchego_version must be %s", avalancheGoVersion)
	}
	if a.RPCChainVMProtocolID != uint32(rpcChainVMProtocol()) {
		return fmt.Errorf("architecture.rpcchainvm_protocol must be %d", rpcChainVMProtocol())
	}
	return nil
}

func (o *AdminOwner) validate(field string) error {
	if len(o.Addresses) == 0 {
		return fmt.Errorf("%s.addresses must list at least one public address", field)
	}
	if len(o.Addresses) > 16 {
		return fmt.Errorf("%s.addresses is bounded to 16 entries", field)
	}
	seen := make(map[string]struct{}, len(o.Addresses))
	for index, address := range o.Addresses {
		entry := fmt.Sprintf("%s.addresses[%d]", field, index)
		if err := refusePlaceholder(entry, address); err != nil {
			return err
		}
		if !pChainAddressRegexp.MatchString(address) {
			return fmt.Errorf("%s must be a public P-chain address", entry)
		}
		if _, duplicate := seen[address]; duplicate {
			return fmt.Errorf("%s repeats an address already listed", entry)
		}
		seen[address] = struct{}{}
	}
	if o.Threshold == 0 || int(o.Threshold) > len(o.Addresses) {
		return fmt.Errorf("%s.threshold must be between 1 and the number of addresses", field)
	}
	return nil
}

func (c *AdminControl) validate() error {
	if err := refusePlaceholder("control.p_chain_controller_address", c.PChainControllerAddress); err != nil {
		return err
	}
	if !pChainAddressRegexp.MatchString(c.PChainControllerAddress) {
		return errors.New("control.p_chain_controller_address must be a public P-chain address")
	}
	if err := c.SubnetOwner.validate("control.subnet_owner"); err != nil {
		return err
	}
	if err := c.RemainingBalanceOwner.validate("control.remaining_balance_owner"); err != nil {
		return err
	}
	return c.DisableOwner.validate("control.disable_owner")
}

func (a *AdminAuthority) validate() error {
	if err := requirePlainHex("authority.formation_public_key", a.FormationPublicKey, ed25519.PublicKeySize); err != nil {
		return err
	}
	seen := map[string]struct{}{a.FormationPublicKey: {}}
	for index, key := range a.ContinuityPublicKeys {
		field := fmt.Sprintf("authority.continuity_public_keys[%d]", index)
		if err := requirePlainHex(field, key, ed25519.PublicKeySize); err != nil {
			return err
		}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("%s duplicates an authority already declared", field)
		}
		seen[key] = struct{}{}
	}
	if a.ExhaustionPlanned && len(a.ContinuityPublicKeys) < 2 {
		return errors.New("authority.continuity_public_keys must list at least two distinct authorities before formation exhaustion is planned")
	}
	if len(a.ContinuityPublicKeys) > 16 {
		return errors.New("authority.continuity_public_keys is bounded to 16 entries")
	}
	return nil
}

func (a *AdminInput) validateValidators() error {
	if len(a.Validators) == 0 {
		return errors.New("validators must list at least one validator: custody is under-specified")
	}
	if len(a.Validators) > 32 {
		return errors.New("validators is bounded to 32 entries")
	}
	if a.Instantiation.Network == networkMainnet && len(a.Validators) < 3 {
		return errors.New("validators must list at least three validators for MAINNET: custody is under-specified")
	}
	nodes := make(map[string]struct{}, len(a.Validators))
	keys := make(map[string]struct{}, len(a.Validators))
	endpoints := make(map[string]struct{}, len(a.Validators))
	for index, validator := range a.Validators {
		field := fmt.Sprintf("validators[%d]", index)
		if err := refusePlaceholder(field+".node_id", validator.NodeID); err != nil {
			return err
		}
		if !nodeIDPattern.MatchString(validator.NodeID) {
			return fmt.Errorf("%s.node_id must be a public NodeID-... identifier", field)
		}
		if err := requirePrefixedHex(field+".bls_public_key", validator.BLSPublicKey, 48); err != nil {
			return err
		}
		if err := requirePrefixedHex(field+".bls_proof_of_possession", validator.BLSProofOfPossession, 96); err != nil {
			return err
		}
		if validator.Weight == 0 {
			return fmt.Errorf("%s.weight must be greater than zero", field)
		}
		if validator.InitialBalanceNanoAVAX == 0 {
			return fmt.Errorf("%s.initial_balance_nanoavax must fund the validator: custody is under-specified", field)
		}
		if validator.RunwayDays == 0 {
			return fmt.Errorf("%s.runway_days must declare the funded continuous-fee runway", field)
		}
		if err := refusePlaceholder(field+".public_p2p_endpoint", validator.PublicP2PEndpoint); err != nil {
			return err
		}
		if !endpointPattern.MatchString(validator.PublicP2PEndpoint) {
			return fmt.Errorf("%s.public_p2p_endpoint must be a reachable host:port", field)
		}
		if validator.HostPlatform != supportedPlatform {
			return fmt.Errorf("%s.host_platform %q is unsupported: this implementation is qualified for %s only", field, validator.HostPlatform, supportedPlatform)
		}
		if validator.KeyCustody != "OPERATOR_HELD_OFFLINE" && validator.KeyCustody != "OPERATOR_HELD_HSM" {
			return fmt.Errorf("%s.key_custody must declare OPERATOR_HELD_OFFLINE or OPERATOR_HELD_HSM: custody is under-specified", field)
		}
		if _, duplicate := nodes[validator.NodeID]; duplicate {
			return fmt.Errorf("%s.node_id duplicates another validator", field)
		}
		nodes[validator.NodeID] = struct{}{}
		if _, duplicate := keys[validator.BLSPublicKey]; duplicate {
			return fmt.Errorf("%s.bls_public_key duplicates another validator", field)
		}
		keys[validator.BLSPublicKey] = struct{}{}
		if _, duplicate := endpoints[validator.PublicP2PEndpoint]; duplicate {
			return fmt.Errorf("%s.public_p2p_endpoint duplicates another validator", field)
		}
		endpoints[validator.PublicP2PEndpoint] = struct{}{}
	}
	return nil
}

func (d *AdminDeclarations) validate() error {
	if d.ValidatorKeyCustody != "OPERATOR_CUSTODY_NEVER_SHARED" {
		return errors.New("declarations.validator_key_custody must state OPERATOR_CUSTODY_NEVER_SHARED")
	}
	if d.BalanceFunding != "OPERATOR_FUNDED_BEFORE_CONVERSION" {
		return errors.New("declarations.balance_funding must state OPERATOR_FUNDED_BEFORE_CONVERSION")
	}
	if d.ConversionAwareness != "ACKNOWLEDGED_IRREVERSIBLE" {
		return errors.New("declarations.conversion_irreversibility_acknowledged must state ACKNOWLEDGED_IRREVERSIBLE")
	}
	if d.ManagerInitialization != "ACKNOWLEDGED_ONE_SHOT" {
		return errors.New("declarations.manager_initialization_is_one_shot_acknowledged must state ACKNOWLEDGED_ONE_SHOT")
	}
	return nil
}
