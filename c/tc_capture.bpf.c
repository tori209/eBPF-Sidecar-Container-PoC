// SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause)
/* Copyright (c) 2022 Hengqi Chen */
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ETH_P_IP  0x0800 /* Internet Protocol packet	*/
/* https://elixir.bootlin.com/linux/v6.8/source/include/uapi/linux/in.h */
#define IPPROTO_IP		0
#define	IPPROTO_ICMP	1
#define IPPROTO_TCP		6
#define	IPPROTO_UDP		17

char __license[] SEC("license") = "GPL";

typedef struct {
	__u64 timestamp_ns;
	__u32 src_ip;
	__u32 dst_ip;
	__u16 sport;
	__u16 dport;
	__u32 size;
	__u8  l4proto;
} l4_metric;

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} log_map SEC(".maps");


SEC("tcx/egress")
int tc_ingress(struct __sk_buff *ctx)
{
	void *data = (void *)(__u64)ctx->data;
	void *data_end = (void *)(__u64)ctx->data_end;
	struct ethhdr *eth;
	struct iphdr *ip;
	l4_metric * event;

	if (ctx->protocol != bpf_htons(ETH_P_IP))
		return TCX_NEXT;

	// Extract Ethernet Header
	// eth + 1 -> end of eth header
	eth = data;
	if ((void *)(eth + 1) > data_end)
		return TCX_NEXT;

	// Extract IP Header
	ip = (struct iphdr *)(eth + 1);
	if ((void *)(ip + 1) > data_end)
		return TCX_NEXT;
	if (ip->protocol != IPPROTO_TCP)
		return TCX_NEXT;

	// Allocate Space of Ringbuf for data
	event = bpf_ringbuf_reserve(&log_map, sizeof(*event), 0);
	if (!event) {
		bpf_printk("Error: ringbuf allocation failed");
		return TCX_NEXT;
	}
	// init memory
	__builtin_memset(event, 0, sizeof(l4_metric));

	event->timestamp_ns = bpf_ktime_get_tai_ns();
	event->src_ip = ip->saddr;
	event->dst_ip = ip->daddr;
	event->size = ctx->len;
	event->l4proto = ip->protocol;

	switch (ip->protocol) {
		case IPPROTO_TCP:
		struct tcphdr *tcph = (void *)((__u8 *)ip + (ip->ihl * 4));
		if ((void *)(tcph + 1) > data_end) {
			bpf_ringbuf_discard(event, 0);
			return XDP_PASS;
		}
		event->sport = bpf_ntohs(tcph->source);
		event->dport = bpf_ntohs(tcph->dest);
		break;

		case IPPROTO_UDP:
		struct udphdr *udph = (void *)((__u8 *)ip + (ip->ihl * 4));
		if ((void *)(udph + 1) > data_end) {
			bpf_ringbuf_discard(event, 0);
			return XDP_PASS;
		}
		event->sport = udph->source;
		event->dport = udph->dest;
		break;

		case IPPROTO_ICMP:
		struct icmphdr *icmph = (void *)((__u8 *)ip + (ip->ihl * 4));
		if ((void *)(icmph + 1) > data_end) {
			bpf_ringbuf_discard(event, 0);
			return XDP_PASS;
		}
		// No port for ICMP
		event->sport = 0;
		event->dport = 0;
		break;

		default:	
		event->sport = 0;
		event->dport = 0;
	}
    bpf_ringbuf_submit(event, 0);
    return XDP_PASS;
}

