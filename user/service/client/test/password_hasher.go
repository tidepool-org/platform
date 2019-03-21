package test

type HashPasswordInput struct {
	UserID   string
	Password string
}

type PasswordHasher struct {
	HashPasswordInvocations int
	HashPasswordInputs      []HashPasswordInput
	HashPasswordStub        func(userID string, password string) string
	HashPasswordOutputs     []string
	HashPasswordOutput      *string
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (p *PasswordHasher) HashPassword(userID string, password string) string {
	p.HashPasswordInvocations++
	p.HashPasswordInputs = append(p.HashPasswordInputs, HashPasswordInput{UserID: userID, Password: password})
	if p.HashPasswordStub != nil {
		return p.HashPasswordStub(userID, password)
	}
	if len(p.HashPasswordOutputs) > 0 {
		output := p.HashPasswordOutputs[0]
		p.HashPasswordOutputs = p.HashPasswordOutputs[1:]
		return output
	}
	if p.HashPasswordOutput != nil {
		return *p.HashPasswordOutput
	}
	panic("HashPassword has no output")
}

func (p *PasswordHasher) AssertOutputsEmpty() {
	if len(p.HashPasswordOutputs) > 0 {
		panic("HashPasswordOutputs is not empty")
	}
}
