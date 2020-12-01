package conditional

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/errors"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
)

type trueBranch struct {
	ctx   *context.Context
	s     token.Stream
	depth int
}

func (b *trueBranch) NextToken() (token.Token, error) {
	if b.depth < 0 {
		return nil, nil
	}
	t, err := b.s.NextToken()
	if token.ErrorOrNil(t, err) {
		return nil, errors.FirstNonNilError(err,
			errors.NewUnexpectedEndOfInputError("reading true branch of if statement"),
		)
	}
	switch classify(t, b.ctx) {
	case ifToken:
		b.depth += 1
	case elseToken:
		if b.depth != 0 {
			break
		}
		err := consumeUntilFi(b.ctx, b.s)
		if err != nil {
			return nil, err
		}
		// we've read the closing fi
		fallthrough
	case fiToken:
		b.depth -= 1
		if b.depth < 0 {
			return nil, nil
		}
	}
	return t, nil
}

func (b *trueBranch) PeekToken() (token.Token, error) {
	if b.depth < 0 {
		return nil, nil
	}
	t, err := b.s.PeekToken()
	if token.ErrorOrNil(t, err) {
		return nil, errors.FirstNonNilError(err,
			errors.NewUnexpectedEndOfInputError("reading true branch of if statement"),
		)
	}
	if b.depth != 0 {
		return t, nil
	}
	switch classify(t, b.ctx) {
	case elseToken:
		b.depth = -1
		return nil, consumeUntilFi(b.ctx, b.s)
	case fiToken:
		_, _ = b.s.NextToken()
		b.depth = -1
		return nil, nil
	}
	return t, nil
}

type falseBranch struct {
	ctx     *context.Context
	s       token.Stream
	depth   int
	started bool
}

func (b *falseBranch) NextToken() (token.Token, error) {
	err := b.ensureStarted()
	if err != nil {
		return nil, err
	}
	if b.depth < 0 {
		return nil, nil
	}
	t, err := b.s.NextToken()
	if token.ErrorOrNil(t, err) {
		return nil, errors.FirstNonNilError(err,
			errors.NewUnexpectedEndOfInputError("reading true branch of if statement"),
		)
	}
	switch classify(t, b.ctx) {
	case ifToken:
		b.depth += 1
	case fiToken:
		b.depth -= 1
		if b.depth < 0 {
			return nil, nil
		}
	}
	return t, nil
}

func (b *falseBranch) PeekToken() (token.Token, error) {
	err := b.ensureStarted()
	if err != nil {
		return nil, err
	}
	if b.depth < 0 {
		return nil, nil
	}
	for {
		t, err := b.s.PeekToken()
		if token.ErrorOrNil(t, err) {
			return nil, errors.FirstNonNilError(err,
				errors.NewUnexpectedEndOfInputError("reading true branch of if statement"),
			)
		}
		switch classify(t, b.ctx) {
		case fiToken:
			if b.depth == 0 {
				_, _ = b.s.NextToken()
				b.depth -= 1
				return nil, nil
			}
		}
		return t, nil
	}
}

func (b *falseBranch) ensureStarted() error {
	if b.started {
		return nil
	}
	lastType, err := consumeUntilElseOrFi(b.ctx, b.s)
	if err != nil {
		return err
	}
	if lastType == fiToken {
		b.depth = -1
	}
	b.started = true
	return nil
}

func consumeUntilFi(ctx *context.Context, s token.Stream) error {
	depth := 0
	for {
		if depth < 0 {
			return nil
		}
		t, err := s.NextToken()
		if token.ErrorOrNil(t, err) {
			return errors.FirstNonNilError(err,
				errors.NewUnexpectedEndOfInputError("skipping the false branch of if statement"),
			)
		}
		switch classify(t, ctx) {
		case ifToken:
			depth += 1
		case fiToken:
			depth -= 1
		}
	}
}

func consumeUntilElseOrFi(ctx *context.Context, s token.Stream) (tokenType, error) {
	depth := 0
	for {
		if depth < 0 {
			return fiToken, nil
		}
		t, err := s.NextToken()
		if token.ErrorOrNil(t, err) {
			return otherToken, errors.FirstNonNilError(err,
				errors.NewUnexpectedEndOfInputError("skipping the true branch of if statement"),
			)
		}
		switch classify(t, ctx) {
		case ifToken:
			depth += 1
		case fiToken:
			depth -= 1
		case elseToken:
			if depth == 0 {
				return elseToken, nil
			}
		}
	}
}

type tokenType int

const (
	ifToken tokenType = iota
	elseToken
	fiToken
	otherToken
)

func classify(t token.Token, ctx *context.Context) tokenType {
	if !t.IsCommand() {
		return otherToken
	}
	cmd, exists := ctx.Expansion.Commands.Get(t.Value())
	switch true {
	case !exists:
		return otherToken
	case IsIfCommand(cmd):
		return ifToken
	case IsFiCommand(cmd):
		return fiToken
	case IsElseCommand(cmd):
		return elseToken
	}
	return otherToken
}
