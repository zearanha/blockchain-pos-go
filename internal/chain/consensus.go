package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"
	"sync"
)

type Validator struct {
	Address string
	Stake   float64
}

// ValidatorSet mantém o registro de todos os validadores e seus stakes.
type ValidatorSet struct {
	mu         sync.Mutex
	validators map[string]*Validator
}

func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		validators: make(map[string]*Validator),
	}
}

// AddStake registra ou aumenta o stake de um endereço.
func (vs *ValidatorSet) AddStake(address string, amount float64) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if v, exists := vs.validators[address]; exists {
		v.Stake += amount
	} else {
		vs.validators[address] = &Validator{Address: address, Stake: amount}
	}
}

// TotalStake retorna a soma de todos os stakes registrados.
func (vs *ValidatorSet) TotalStake() float64 {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	total := 0.0
	for _, v := range vs.validators {
		total += v.Stake
	}
	return total
}

func (vs *ValidatorSet) StakeOf(address string) (float64, bool) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	v, exists := vs.validators[address]
	if !exists {
		return 0, false
	}
	return v.Stake, true
}

func (vs *ValidatorSet) Slash(address string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if v, exists := vs.validators[address]; exists {
		v.Stake = 0
		return
	}

	vs.validators[address] = &Validator{Address: address, Stake: 0}
}

// SelectValidator escolhe um validador de forma determinística, ponderada pelo stake.
// A seed (normalmente o hash do bloco anterior) garante que todos os nós,
// com o mesmo ValidatorSet, cheguem ao mesmo resultado.
func (vs *ValidatorSet) SelectValidator(seed string) (string, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if len(vs.validators) == 0 {
		return "", errors.New("nenhum validador registrado")
	}

	totalStake := 0.0
	for _, v := range vs.validators {
		totalStake += v.Stake
	}
	if totalStake == 0 {
		return "", errors.New("stake total é zero")
	}

	addresses := make([]string, 0, len(vs.validators))
	for addr := range vs.validators {
		addresses = append(addresses, addr)
	}
	sort.Strings(addresses)

	point := seededRandom(seed, totalStake)

	cumulative := 0.0
	for _, addr := range addresses {
		cumulative += vs.validators[addr].Stake
		if point < cumulative {
			return addr, nil
		}
	}

	return addresses[len(addresses)-1], nil
}

func seededRandom(seed string, max float64) float64 {
	hash := sha256.Sum256([]byte(seed))
	n := binary.BigEndian.Uint64(hash[:8])

	fraction := float64(n) / float64(^uint64(0))
	return fraction * max
}
