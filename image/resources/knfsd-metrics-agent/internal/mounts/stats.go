package mounts

import (
	"time"

	"github.com/prometheus/procfs"
)

type summary struct {
	age       time.Duration
	bytes     procfs.NFSBytesStats
	events    procfs.NFSEventsStats
	transport procfs.NFSTransportStats
	read      procfs.NFSOperationStats
	write     procfs.NFSOperationStats
}

func newSummary(new *procfs.MountStatsNFS) summary {
	read, _ := find(new.Operations, "READ")
	write, _ := find(new.Operations, "WRITE")
	return summary{
		age:       new.Age,
		bytes:     new.Bytes,
		events:    new.Events,
		transport: new.Transport,
		read:      read,
		write:     write,
	}
}

func diffSummary(new, old summary) summary {
	return summary{
		age:       new.age - old.age,
		bytes:     diffBytes(new.bytes, old.bytes),
		events:    diffEvents(new.events, old.events),
		transport: diffTransport(new.transport, old.transport),
		read:      diffOperation(new.read, old.read),
		write:     diffOperation(new.write, old.write),
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

func sub(x, y uint64) uint64 {
	if x >= y {
		return x - y
	} else {
		// prevent underflow, clamp to zero
		return 0
	}
}
