// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#ifndef PIGEON_PATHSWITCH_GEN_H
#define PIGEON_PATHSWITCH_GEN_H

#include <stdbool.h>
#include <stdint.h>

// PathSwitch backend states.
typedef enum {
	PIGEON_BACKEND_RELAY_CONNECTED = 0,
	PIGEON_BACKEND_LAN_OFFERED,
	PIGEON_BACKEND_LAN_ACTIVE,
	PIGEON_BACKEND_RELAY_BACKOFF,
	PIGEON_BACKEND_LAN_DEGRADED,
	PIGEON_BACKEND_STATE_COUNT
} pigeon_backend_state;

// PathSwitch client states.
typedef enum {
	PIGEON_CLIENT_RELAY_CONNECTED = 0,
	PIGEON_CLIENT_LAN_CONNECTING,
	PIGEON_CLIENT_LAN_VERIFYING,
	PIGEON_CLIENT_LAN_ACTIVE,
	PIGEON_CLIENT_RELAY_FALLBACK,
	PIGEON_CLIENT_STATE_COUNT
} pigeon_client_state;

// PathSwitch relay states.
typedef enum {
	PIGEON_RELAY_IDLE = 0,
	PIGEON_RELAY_BACKEND_REGISTERED,
	PIGEON_RELAY_BRIDGED,
	PIGEON_RELAY_STATE_COUNT
} pigeon_relay_state;

// PathSwitch message types.
typedef enum {
	PIGEON_PATHSWITCH_MSG_LAN_OFFER = 0,
	PIGEON_PATHSWITCH_MSG_LAN_VERIFY,
	PIGEON_PATHSWITCH_MSG_LAN_CONFIRM,
	PIGEON_PATHSWITCH_MSG_PATH_PING,
	PIGEON_PATHSWITCH_MSG_PATH_PONG,
	PIGEON_PATHSWITCH_MSG_RELAY_RESUME,
	PIGEON_PATHSWITCH_MSG_RELAY_RESUMED,
	PIGEON_PATHSWITCH_MSG_COUNT
} path_switch_msg_type;

// PathSwitch guards.
typedef enum {
	PIGEON_PATHSWITCH_GUARD_CHALLENGE_VALID = 0,
	PIGEON_PATHSWITCH_GUARD_CHALLENGE_INVALID,
	PIGEON_PATHSWITCH_GUARD_LAN_ENABLED,
	PIGEON_PATHSWITCH_GUARD_LAN_DISABLED,
	PIGEON_PATHSWITCH_GUARD_LAN_SERVER_AVAILABLE,
	PIGEON_PATHSWITCH_GUARD_UNDER_MAX_FAILURES,
	PIGEON_PATHSWITCH_GUARD_AT_MAX_FAILURES,
	PIGEON_PATHSWITCH_GUARD_COUNT
} path_switch_guard_id;

// PathSwitch actions.
typedef enum {
	PIGEON_PATHSWITCH_ACTION_ACTIVATE_LAN = 0,
	PIGEON_PATHSWITCH_ACTION_RESET_FAILURES,
	PIGEON_PATHSWITCH_ACTION_FALLBACK_TO_RELAY,
	PIGEON_PATHSWITCH_ACTION_DIAL_LAN,
	PIGEON_PATHSWITCH_ACTION_BRIDGE_STREAMS,
	PIGEON_PATHSWITCH_ACTION_UNBRIDGE,
	PIGEON_PATHSWITCH_ACTION_REBRIDGE_STREAMS,
	PIGEON_PATHSWITCH_ACTION_COUNT
} path_switch_action_id;

// PathSwitch events.
typedef enum {
	PIGEON_PATHSWITCH_EVENT_LAN_SERVER_READY = 0,
	PIGEON_PATHSWITCH_EVENT_OFFER_TIMEOUT,
	PIGEON_PATHSWITCH_EVENT_PING_TICK,
	PIGEON_PATHSWITCH_EVENT_PING_TIMEOUT,
	PIGEON_PATHSWITCH_EVENT_BACKOFF_EXPIRED,
	PIGEON_PATHSWITCH_EVENT_LAN_SERVER_CHANGED,
	PIGEON_PATHSWITCH_EVENT_READVERTISE_TICK,
	PIGEON_PATHSWITCH_EVENT_LAN_DIAL_OK,
	PIGEON_PATHSWITCH_EVENT_LAN_DIAL_FAILED,
	PIGEON_PATHSWITCH_EVENT_VERIFY_TIMEOUT,
	PIGEON_PATHSWITCH_EVENT_LAN_ERROR,
	PIGEON_PATHSWITCH_EVENT_RELAY_OK,
	PIGEON_PATHSWITCH_EVENT_BACKEND_REGISTER,
	PIGEON_PATHSWITCH_EVENT_CLIENT_CONNECT,
	PIGEON_PATHSWITCH_EVENT_CLIENT_DISCONNECT,
	PIGEON_PATHSWITCH_EVENT_BACKEND_DISCONNECT,
	PIGEON_PATHSWITCH_EVENT_RECV_LAN_VERIFY,
	PIGEON_PATHSWITCH_EVENT_RECV_PATH_PONG,
	PIGEON_PATHSWITCH_EVENT_RECV_LAN_OFFER,
	PIGEON_PATHSWITCH_EVENT_RECV_LAN_CONFIRM,
	PIGEON_PATHSWITCH_EVENT_RECV_PATH_PING,
	PIGEON_PATHSWITCH_EVENT_RECV_RELAY_RESUME,
	PIGEON_PATHSWITCH_EVENT_COUNT
} path_switch_event_id;

// Guard and action callback types.
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// PathSwitch backend state machine.
typedef struct {
	pigeon_backend_state state;
	int ping_failures; // consecutive failed pings on the direct path
	int backoff_level; // current exponential backoff level (0 = no backoff)
	const char * active_path; // "relay" or "lan" — which path carries application traffic
	const char * dispatcher_path; // which path the datagram dispatcher reads from ("relay", "lan", "none")
	const char * monitor_target; // which path the health monitor pings ("lan", "none")
	const char * lan_signal; // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)
	pigeon_guard_fn guards[PIGEON_PATHSWITCH_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_PATHSWITCH_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_backend_machine;

void pigeon_backend_machine_init(pigeon_backend_machine *m);
int  pigeon_backend_handle_message(pigeon_backend_machine *m, path_switch_msg_type msg);
int  pigeon_backend_step(pigeon_backend_machine *m, path_switch_event_id event);

// PathSwitch client state machine.
typedef struct {
	pigeon_client_state state;
	const char * active_path; // "relay" or "lan" — which path carries application traffic
	const char * dispatcher_path; // which path the datagram dispatcher reads from ("relay", "lan", "none")
	const char * lan_signal; // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)
	pigeon_guard_fn guards[PIGEON_PATHSWITCH_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_PATHSWITCH_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_client_machine;

void pigeon_client_machine_init(pigeon_client_machine *m);
int  pigeon_client_handle_message(pigeon_client_machine *m, path_switch_msg_type msg);
int  pigeon_client_step(pigeon_client_machine *m, path_switch_event_id event);

// PathSwitch relay state machine.
typedef struct {
	pigeon_relay_state state;
	const char * relay_bridge; // relay bridge state ("active" = bridging, "idle" = backend registered but no client)
	pigeon_guard_fn guards[PIGEON_PATHSWITCH_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_PATHSWITCH_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_relay_machine;

void pigeon_relay_machine_init(pigeon_relay_machine *m);
int  pigeon_relay_handle_message(pigeon_relay_machine *m, path_switch_msg_type msg);
int  pigeon_relay_step(pigeon_relay_machine *m, path_switch_event_id event);

#endif // PIGEON_PATHSWITCH_GEN_H
