package jailconf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

// Parse reads jail.conf-format data and returns a map of jail name to Config.
func Parse(data []byte) (map[string]jail.Config, error) {
	result := make(map[string]jail.Config)
	p := &parser{input: string(data)}

	for p.skipWhitespaceAndComments(); p.pos < len(p.input); p.skipWhitespaceAndComments() {
		name, err := p.readJailName()
		if err != nil {
			return nil, err
		}

		cfg, err := p.readBlock()
		if err != nil {
			return nil, fmt.Errorf("jail %q: %w", name, err)
		}

		cfg.Name = name
		result[name] = cfg
	}

	return result, nil
}

type parser struct {
	input string
	pos   int
	line  int
}

func (p *parser) readJailName() (string, error) {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '{' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			break
		}
		p.pos++
	}
	name := strings.TrimSpace(p.input[start:p.pos])
	if name == "" {
		return "", fmt.Errorf("line %d: expected jail name", p.line)
	}
	return name, nil
}

func (p *parser) readBlock() (jail.Config, error) {
	var cfg jail.Config

	p.skipWhitespaceAndComments()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return cfg, fmt.Errorf("line %d: expected '{'", p.line)
	}
	p.pos++ // skip '{'

	for {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			return cfg, fmt.Errorf("line %d: unexpected end of input, expected '}'", p.line)
		}
		if p.input[p.pos] == '}' {
			p.pos++
			break
		}

		key, value, isBool, err := p.readStatement()
		if err != nil {
			return cfg, err
		}

		if err := applyParam(&cfg, key, value, isBool); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

func (p *parser) readStatement() (key, value string, isBool bool, err error) {
	// Read key
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '=' || ch == ';' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			break
		}
		p.pos++
	}
	key = strings.TrimSpace(p.input[start:p.pos])
	if key == "" {
		err = fmt.Errorf("line %d: expected parameter name", p.line)
		return
	}

	p.skipSpaces()

	if p.pos >= len(p.input) {
		err = fmt.Errorf("line %d: unexpected end of input", p.line)
		return
	}

	if p.input[p.pos] == ';' {
		// Boolean parameter: "key;"
		p.pos++
		isBool = true
		return
	}

	if p.input[p.pos] == '=' {
		p.pos++ // skip '='
		p.skipSpaces()

		if p.pos < len(p.input) && p.input[p.pos] == '"' {
			// Quoted value
			value, err = p.readQuotedString()
			if err != nil {
				return
			}
		} else {
			// Unquoted value
			vstart := p.pos
			for p.pos < len(p.input) && p.input[p.pos] != ';' && p.input[p.pos] != '\n' {
				p.pos++
			}
			value = strings.TrimSpace(p.input[vstart:p.pos])
		}

		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ';' {
			p.pos++
		}
		return
	}

	// Bare word followed by newline or space (treat as bool)
	isBool = true
	// Skip to semicolon if present
	for p.pos < len(p.input) && p.input[p.pos] != ';' && p.input[p.pos] != '\n' && p.input[p.pos] != '}' {
		p.pos++
	}
	if p.pos < len(p.input) && p.input[p.pos] == ';' {
		p.pos++
	}
	return
}

func (p *parser) readQuotedString() (string, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '"' {
		return "", fmt.Errorf("line %d: expected '\"'", p.line)
	}
	p.pos++ // skip opening quote

	var b strings.Builder
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '\\' && p.pos+1 < len(p.input) {
			p.pos++
			b.WriteByte(p.input[p.pos])
			p.pos++
			continue
		}
		if ch == '"' {
			p.pos++
			return b.String(), nil
		}
		if ch == '\n' {
			p.line++
		}
		b.WriteByte(ch)
		p.pos++
	}
	return "", fmt.Errorf("line %d: unterminated string", p.line)
}

func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '\n' {
			p.line++
			p.pos++
		} else if ch == ' ' || ch == '\t' || ch == '\r' {
			p.pos++
		} else if ch == '#' {
			// Skip to end of line
			for p.pos < len(p.input) && p.input[p.pos] != '\n' {
				p.pos++
			}
		} else {
			break
		}
	}
}

func (p *parser) skipSpaces() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t') {
		p.pos++
	}
}

func applyParam(cfg *jail.Config, key, value string, isBool bool) error {
	switch key {
	case "path":
		cfg.Path = value
	case "host.hostname":
		cfg.Hostname = value
	case "persist":
		cfg.Persist = true
	case "vnet":
		cfg.VNET = true
	case "vnet.interface":
		cfg.VNETInterface = append(cfg.VNETInterface, value)
	case "ip4.addr":
		for _, addr := range strings.Split(value, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				cfg.IP4 = append(cfg.IP4, addr)
			}
		}
	case "ip6.addr":
		for _, addr := range strings.Split(value, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				cfg.IP6 = append(cfg.IP6, addr)
			}
		}
	case "mount.devfs":
		cfg.MountDevFS = true
	case "devfs_ruleset":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid devfs_ruleset %q: %w", value, err)
		}
		cfg.DevFSRuleset = &n
	case "children.max":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid children.max %q: %w", value, err)
		}
		cfg.Limits.ChildrenMax = &n
	case "allow.raw_sockets":
		cfg.Allow.RawSockets = true
	case "allow.mount":
		cfg.Allow.Mount = true
	case "allow.mount.devfs":
		cfg.Allow.MountDevFS = true
	case "allow.mount.nullfs":
		cfg.Allow.MountNullFS = true
	case "allow.mount.procfs":
		cfg.Allow.MountProcFS = true
	case "allow.mount.tmpfs":
		cfg.Allow.MountTmpFS = true
	case "allow.sysvipc":
		cfg.Allow.SysVIPC = true
	case "allow.set_hostname":
		cfg.Allow.SetHostname = true
	case "allow.chflags":
		cfg.Allow.Chflags = true
	case "allow.socket_af":
		cfg.Allow.SocketAF = true
	default:
		// Store unknown params in RawParams
		if cfg.RawParams == nil {
			cfg.RawParams = make(map[string]string)
		}
		if isBool {
			cfg.RawParams[key] = ""
		} else {
			cfg.RawParams[key] = value
		}
	}
	return nil
}
