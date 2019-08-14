/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package gossip

import (
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log2"
	"github.com/bbva/qed/metrics"
	"github.com/coocood/freecache"
)

type AgentOptionF func(*Agent) error

func configToOptions(conf *Config) ([]AgentOptionF, error) {
	var options []AgentOptionF

	if conf == nil {
		return nil, nil
	}

	options = []AgentOptionF{
		SetNodeName(conf.NodeName),
		SetRole(conf.Role),
		SetBindAddr(conf.BindAddr),
		SetAdvertiseAddr(conf.AdvertiseAddr),
		SetLeaveOnTerm(conf.LeaveOnTerm),
		SetStartJoin(conf.StartJoin),
		SetEnableCompression(conf.EnableCompression),
		SetBroadcastTimeout(conf.BroadcastTimeout),
		SetLeavePropagateDelay(conf.LeavePropagateDelay),
		SetTimeoutQueues(conf.TimeoutQueues),
		SetProcessInterval(conf.ProcessInterval),
		SetMetricsServer(conf.MetricsAddr),
		SetCache(conf.CacheSize),
		SetTimeoutQueues(conf.TimeoutQueues),
	}

	return options, nil
}

func SetNodeName(name string) AgentOptionF {
	return func(a *Agent) error {
		a.config.NodeName = name
		return nil
	}
}

func SetRole(role string) AgentOptionF {
	return func(a *Agent) error {
		a.config.Role = role
		return nil
	}
}

func SetBindAddr(addr string) AgentOptionF {
	return func(a *Agent) error {
		a.config.BindAddr = addr
		return nil
	}
}

func SetAdvertiseAddr(addr string) AgentOptionF {
	return func(a *Agent) error {
		a.config.AdvertiseAddr = addr
		return nil
	}
}

func SetLeaveOnTerm(leave bool) AgentOptionF {
	return func(a *Agent) error {
		a.config.LeaveOnTerm = leave
		return nil
	}
}

func SetStartJoin(addrs []string) AgentOptionF {
	return func(a *Agent) error {
		a.config.StartJoin = addrs
		return nil
	}
}

func SetEnableCompression(enabled bool) AgentOptionF {
	return func(a *Agent) error {
		a.config.EnableCompression = enabled
		return nil
	}
}

func SetBroadcastTimeout(timeout time.Duration) AgentOptionF {
	return func(a *Agent) error {
		a.config.BroadcastTimeout = timeout
		return nil
	}
}

func SetLeavePropagateDelay(delay time.Duration) AgentOptionF {
	return func(a *Agent) error {
		a.config.LeavePropagateDelay = delay
		return nil
	}
}

func SetTimeoutQueues(timeout time.Duration) AgentOptionF {
	return func(a *Agent) error {
		a.timeout = time.NewTicker(timeout)
		return nil
	}
}

func SetProcessInterval(interval time.Duration) AgentOptionF {
	return func(a *Agent) error {
		return nil
	}
}

func SetMetricsServer(addr string) AgentOptionF {
	return func(a *Agent) error {
		if addr != "" {
			a.metrics = metrics.NewServer(addr)
		}
		return nil
	}
}

func SetTasksManager(tm TasksManager) AgentOptionF {
	return func(a *Agent) error {
		a.Tasks = tm
		return nil
	}
}

func SetQEDClient(qed *client.HTTPClient) AgentOptionF {
	return func(a *Agent) error {
		a.Qed = qed
		return nil
	}
}

func SetSnapshotStore(store SnapshotStore) AgentOptionF {
	return func(a *Agent) error {
		a.SnapshotStore = store
		return nil
	}
}

func SetNotifier(n Notifier) AgentOptionF {
	return func(a *Agent) error {
		a.Notifier = n
		return nil
	}
}

func SetLogger(l log2.Logger) AgentOptionF {
	return func(a *Agent) error {
		a.log = l
		if a.log == nil {
			a.log = log2.L()
		}
		return nil
	}
}

// export GOGC variable to make GC to collect memory
// adecuately if the cache is too big
func SetCache(size int) AgentOptionF {
	return func(a *Agent) error {
		a.Cache = freecache.NewCache(int(size))
		return nil
	}
}

func SetProcessors(p map[string]Processor) AgentOptionF {
	return func(a *Agent) error {
		for _, p := range p {
			a.RegisterMetrics(p.Metrics())
		}
		a.processors = p
		return nil
	}
}
