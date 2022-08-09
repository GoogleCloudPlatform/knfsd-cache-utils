/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package opt

import (
	"flag"
	"fmt"
	"os"
)

type OptSet struct {
	LookupEnv func(string) (string, bool)

	flags *flag.FlagSet
	env   map[string]string // maps flag name to environment variable name
}

func NewOptSet(f *flag.FlagSet) *OptSet {
	return &OptSet{flags: f}
}

func (opts *OptSet) Parse(arguments []string) error {
	flags := opts.flags

	if opts.LookupEnv == nil {
		opts.LookupEnv = os.LookupEnv
	}

	err := flags.Parse(arguments)
	if err != nil {
		return err
	}

	// copy the environment map
	env := make(map[string]string)
	for k, v := range opts.env {
		env[k] = v
	}

	// remove arguments that have already been explicitly set by the command line flags
	flags.Visit(func(f *flag.Flag) {
		delete(env, f.Name)
	})

	// set arguments from the environment
	for f, e := range env {
		err := opts.setFromEnv(f, e)
		if err != nil {
			switch flags.ErrorHandling() {
			case flag.ContinueOnError:
				return err
			case flag.ExitOnError:
				fmt.Fprintln(flags.Output(), err)
				flags.Usage()
				os.Exit(2)
			case flag.PanicOnError:
				panic(err)
			}
		}
	}

	return nil
}

func (opts *OptSet) setFromEnv(flag, env string) error {
	val, present := opts.LookupEnv(env)
	if present && val != "" {
		err := opts.flags.Set(flag, val)
		if err != nil {
			return fmt.Errorf("invalid value \"%s\" for %s: %w", val, env, err)
		}
	}
	return nil
}

func (opts *OptSet) add(flag, env string) {
	if env == "" {
		return
	}

	_, exists := opts.env[flag]
	if exists {
		panic(fmt.Sprintf("environment variable redefined: %s", env))
	}

	if opts.env == nil {
		opts.env = make(map[string]string)
	}
	opts.env[flag] = env
}

func (opts *OptSet) StringVar(p *string, name, env, usage string) {
	opts.flags.StringVar(p, name, "", formatUsage(usage, env))
	opts.add(name, env)
}

func (opts *OptSet) BoolVar(p *bool, name, env, usage string) {
	opts.flags.BoolVar(p, name, false, formatUsage(usage, env))
	opts.add(name, env)
}

func formatUsage(usage, env string) string {
	if env == "" {
		return usage
	} else {
		return fmt.Sprintf("%s [%s]", usage, env)
	}
}
