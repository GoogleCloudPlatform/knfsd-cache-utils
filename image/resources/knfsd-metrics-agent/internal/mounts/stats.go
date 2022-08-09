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

package mounts

import (
	"math"
	"time"

	"github.com/prometheus/procfs"
)

type summary struct {
	age        time.Duration
	bytes      procfs.NFSBytesStats
	events     procfs.NFSEventsStats
	transport  procfs.NFSTransportStats
	operations map[string]procfs.NFSOperationStats
}

func newSummary(new *procfs.MountStatsNFS) summary {
	operations := make(map[string]procfs.NFSOperationStats, len(new.Operations))
	for _, op := range new.Operations {
		operations[op.Operation] = op
	}
	return summary{
		age:        new.Age,
		bytes:      new.Bytes,
		events:     new.Events,
		transport:  new.Transport,
		operations: operations,
	}
}

func addSummary(new, old summary) summary {
	age := new.age
	if age < old.age {
		age = old.age
	}

	return summary{
		age:        age,
		bytes:      addBytes(new.bytes, old.bytes),
		events:     addEvents(new.events, old.events),
		transport:  addTransport(new.transport, old.transport),
		operations: addOperations(new.operations, old.operations),
	}
}

func addBytes(new, old procfs.NFSBytesStats) procfs.NFSBytesStats {
	return procfs.NFSBytesStats{
		Read:        add(new.Read, old.Read),
		Write:       add(new.Write, old.Write),
		DirectRead:  add(new.DirectRead, old.DirectRead),
		DirectWrite: add(new.DirectWrite, old.DirectWrite),
		ReadTotal:   add(new.ReadTotal, old.ReadTotal),
		WriteTotal:  add(new.WriteTotal, old.WriteTotal),
		ReadPages:   add(new.ReadPages, old.ReadPages),
		WritePages:  add(new.WritePages, old.WritePages),
	}
}

func addEvents(new, old procfs.NFSEventsStats) procfs.NFSEventsStats {
	return procfs.NFSEventsStats{
		InodeRevalidate:     add(new.InodeRevalidate, old.InodeRevalidate),
		DnodeRevalidate:     add(new.DnodeRevalidate, old.DnodeRevalidate),
		DataInvalidate:      add(new.DataInvalidate, old.DataInvalidate),
		AttributeInvalidate: add(new.AttributeInvalidate, old.AttributeInvalidate),
		VFSOpen:             add(new.VFSOpen, old.VFSOpen),
		VFSLookup:           add(new.VFSLookup, old.VFSLookup),
		VFSAccess:           add(new.VFSAccess, old.VFSAccess),
		VFSUpdatePage:       add(new.VFSUpdatePage, old.VFSUpdatePage),
		VFSReadPage:         add(new.VFSReadPage, old.VFSReadPage),
		VFSReadPages:        add(new.VFSReadPages, old.VFSReadPages),
		VFSWritePage:        add(new.VFSWritePage, old.VFSWritePage),
		VFSWritePages:       add(new.VFSWritePages, old.VFSWritePages),
		VFSGetdents:         add(new.VFSGetdents, old.VFSGetdents),
		VFSSetattr:          add(new.VFSSetattr, old.VFSSetattr),
		VFSFlush:            add(new.VFSFlush, old.VFSFlush),
		VFSFsync:            add(new.VFSFsync, old.VFSFsync),
		VFSLock:             add(new.VFSLock, old.VFSLock),
		VFSFileRelease:      add(new.VFSFileRelease, old.VFSFileRelease),
		CongestionWait:      add(new.CongestionWait, old.CongestionWait),
		Truncation:          add(new.Truncation, old.Truncation),
		WriteExtension:      add(new.WriteExtension, old.WriteExtension),
		SillyRename:         add(new.SillyRename, old.SillyRename),
		ShortRead:           add(new.ShortRead, old.ShortRead),
		ShortWrite:          add(new.ShortWrite, old.ShortWrite),
		JukeboxDelay:        add(new.JukeboxDelay, old.JukeboxDelay),
		PNFSRead:            add(new.PNFSRead, old.PNFSRead),
		PNFSWrite:           add(new.PNFSWrite, old.PNFSWrite),
	}
}

func addTransport(new, old procfs.NFSTransportStats) procfs.NFSTransportStats {
	return procfs.NFSTransportStats{
		Protocol:                 new.Protocol,
		Port:                     new.Port,
		Bind:                     add(new.Bind, old.Bind),
		Connect:                  add(new.Connect, old.Connect),
		ConnectIdleTime:          add(new.ConnectIdleTime, old.ConnectIdleTime),
		IdleTimeSeconds:          add(new.IdleTimeSeconds, old.IdleTimeSeconds),
		Sends:                    add(new.Sends, old.Sends),
		Receives:                 add(new.Receives, old.Receives),
		BadTransactionIDs:        add(new.BadTransactionIDs, old.BadTransactionIDs),
		CumulativeActiveRequests: add(new.CumulativeActiveRequests, old.CumulativeActiveRequests),
		CumulativeBacklog:        add(new.CumulativeBacklog, old.CumulativeBacklog),
		MaximumRPCSlotsUsed:      add(new.MaximumRPCSlotsUsed, old.MaximumRPCSlotsUsed),
		CumulativeSendingQueue:   add(new.CumulativeSendingQueue, old.CumulativeSendingQueue),
		CumulativePendingQueue:   add(new.CumulativePendingQueue, old.CumulativePendingQueue),
	}
}

func addOperations(new, old map[string]procfs.NFSOperationStats) map[string]procfs.NFSOperationStats {
	cap := len(new)
	if x := len(old); x > cap {
		cap = x
	}

	sum := make(map[string]procfs.NFSOperationStats)
	for key, value := range new {
		sum[key] = value
	}

	for key, value := range old {
		sum[key] = addOperation(key, sum[key], value)
	}

	return sum
}

func addOperation(operation string, new, old procfs.NFSOperationStats) procfs.NFSOperationStats {
	return procfs.NFSOperationStats{
		Operation:                           operation,
		Requests:                            add(new.Requests, old.Requests),
		Transmissions:                       add(new.Transmissions, old.Transmissions),
		MajorTimeouts:                       add(new.MajorTimeouts, old.MajorTimeouts),
		BytesSent:                           add(new.BytesSent, old.BytesSent),
		BytesReceived:                       add(new.BytesReceived, old.BytesReceived),
		CumulativeQueueMilliseconds:         add(new.CumulativeQueueMilliseconds, old.CumulativeQueueMilliseconds),
		CumulativeTotalResponseMilliseconds: add(new.CumulativeTotalResponseMilliseconds, old.CumulativeTotalResponseMilliseconds),
		CumulativeTotalRequestMilliseconds:  add(new.CumulativeTotalRequestMilliseconds, old.CumulativeTotalRequestMilliseconds),
		Errors:                              add(new.Errors, old.Errors),
	}
}

func diffSummary(new, old summary) summary {
	return summary{
		age:        new.age - old.age,
		bytes:      diffBytes(new.bytes, old.bytes),
		events:     diffEvents(new.events, old.events),
		transport:  diffTransport(new.transport, old.transport),
		operations: diffOperations(new.operations, old.operations),
	}
}

func diffBytes(new, old procfs.NFSBytesStats) procfs.NFSBytesStats {
	return procfs.NFSBytesStats{
		Read:        sub(new.Read, old.Read),
		Write:       sub(new.Write, old.Write),
		DirectRead:  sub(new.DirectRead, old.DirectRead),
		DirectWrite: sub(new.DirectWrite, old.DirectWrite),
		ReadTotal:   sub(new.ReadTotal, old.ReadTotal),
		WriteTotal:  sub(new.WriteTotal, old.WriteTotal),
		ReadPages:   sub(new.ReadPages, old.ReadPages),
		WritePages:  sub(new.WritePages, old.WritePages),
	}
}

func diffEvents(new, old procfs.NFSEventsStats) procfs.NFSEventsStats {
	return procfs.NFSEventsStats{
		InodeRevalidate:     sub(new.InodeRevalidate, old.InodeRevalidate),
		DnodeRevalidate:     sub(new.DnodeRevalidate, old.DnodeRevalidate),
		DataInvalidate:      sub(new.DataInvalidate, old.DataInvalidate),
		AttributeInvalidate: sub(new.AttributeInvalidate, old.AttributeInvalidate),
		VFSOpen:             sub(new.VFSOpen, old.VFSOpen),
		VFSLookup:           sub(new.VFSLookup, old.VFSLookup),
		VFSAccess:           sub(new.VFSAccess, old.VFSAccess),
		VFSUpdatePage:       sub(new.VFSUpdatePage, old.VFSUpdatePage),
		VFSReadPage:         sub(new.VFSReadPage, old.VFSReadPage),
		VFSReadPages:        sub(new.VFSReadPages, old.VFSReadPages),
		VFSWritePage:        sub(new.VFSWritePage, old.VFSWritePage),
		VFSWritePages:       sub(new.VFSWritePages, old.VFSWritePages),
		VFSGetdents:         sub(new.VFSGetdents, old.VFSGetdents),
		VFSSetattr:          sub(new.VFSSetattr, old.VFSSetattr),
		VFSFlush:            sub(new.VFSFlush, old.VFSFlush),
		VFSFsync:            sub(new.VFSFsync, old.VFSFsync),
		VFSLock:             sub(new.VFSLock, old.VFSLock),
		VFSFileRelease:      sub(new.VFSFileRelease, old.VFSFileRelease),
		CongestionWait:      sub(new.CongestionWait, old.CongestionWait),
		Truncation:          sub(new.Truncation, old.Truncation),
		WriteExtension:      sub(new.WriteExtension, old.WriteExtension),
		SillyRename:         sub(new.SillyRename, old.SillyRename),
		ShortRead:           sub(new.ShortRead, old.ShortRead),
		ShortWrite:          sub(new.ShortWrite, old.ShortWrite),
		JukeboxDelay:        sub(new.JukeboxDelay, old.JukeboxDelay),
		PNFSRead:            sub(new.PNFSRead, old.PNFSRead),
		PNFSWrite:           sub(new.PNFSWrite, old.PNFSWrite),
	}
}

func diffTransport(new, old procfs.NFSTransportStats) procfs.NFSTransportStats {
	return procfs.NFSTransportStats{
		Protocol:                 new.Protocol,
		Port:                     new.Port,
		Bind:                     sub(new.Bind, old.Bind),
		Connect:                  sub(new.Connect, old.Connect),
		ConnectIdleTime:          sub(new.ConnectIdleTime, old.ConnectIdleTime),
		IdleTimeSeconds:          sub(new.IdleTimeSeconds, old.IdleTimeSeconds),
		Sends:                    sub(new.Sends, old.Sends),
		Receives:                 sub(new.Receives, old.Receives),
		BadTransactionIDs:        sub(new.BadTransactionIDs, old.BadTransactionIDs),
		CumulativeActiveRequests: sub(new.CumulativeActiveRequests, old.CumulativeActiveRequests),
		CumulativeBacklog:        sub(new.CumulativeBacklog, old.CumulativeBacklog),
		MaximumRPCSlotsUsed:      sub(new.MaximumRPCSlotsUsed, old.MaximumRPCSlotsUsed),
		CumulativeSendingQueue:   sub(new.CumulativeSendingQueue, old.CumulativeSendingQueue),
		CumulativePendingQueue:   sub(new.CumulativePendingQueue, old.CumulativePendingQueue),
	}
}

func diffOperations(new, old map[string]procfs.NFSOperationStats) map[string]procfs.NFSOperationStats {
	diff := make(map[string]procfs.NFSOperationStats, len(new))
	for key, newOp := range new {
		if oldOp, found := old[key]; found {
			diff[key] = diffOperation(newOp, oldOp)
		} else {
			// Both maps should be identical as the mountstats always contain
			// the same list of operations. In case they are different, we only
			// care about the operations in the new list.
			diff[key] = newOp
		}
	}
	return diff
}

func diffOperation(new, old procfs.NFSOperationStats) procfs.NFSOperationStats {
	return procfs.NFSOperationStats{
		Operation:                           new.Operation,
		Requests:                            sub(new.Requests, old.Requests),
		Transmissions:                       sub(new.Transmissions, old.Transmissions),
		MajorTimeouts:                       sub(new.MajorTimeouts, old.MajorTimeouts),
		BytesSent:                           sub(new.BytesSent, old.BytesSent),
		BytesReceived:                       sub(new.BytesReceived, old.BytesReceived),
		CumulativeQueueMilliseconds:         sub(new.CumulativeQueueMilliseconds, old.CumulativeQueueMilliseconds),
		CumulativeTotalResponseMilliseconds: sub(new.CumulativeTotalResponseMilliseconds, old.CumulativeTotalResponseMilliseconds),
		CumulativeTotalRequestMilliseconds:  sub(new.CumulativeTotalRequestMilliseconds, old.CumulativeTotalRequestMilliseconds),
		Errors:                              sub(new.Errors, old.Errors),
	}
}

func add(x, y uint64) uint64 {
	cap := math.MaxUint64 - x
	if y >= cap {
		// prevent overflow, clamp to MaxUint64
		return math.MaxUint64
	} else {
		return x + y
	}
}

func sub(x, y uint64) uint64 {
	if x > y {
		return x - y
	} else {
		// prevent underflow, clamp to zero
		return 0
	}
}
