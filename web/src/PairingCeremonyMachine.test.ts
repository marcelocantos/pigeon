// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Unit tests for the regenerated pairing-ceremony state machines.
// These exercise the spec produced by `cmd/protogen protocol/pairing.yaml`
// and pin the TypeScript outputs to the same shape used by Go, Swift,
// Kotlin, C, and TLA+.

import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  PairingCeremonyAcceptorState,
  PairingCeremonyInitiatorState,
  PairingCeremonyProtocol,
  PairingCeremonyAcceptorMachine,
  PairingCeremonyInitiatorMachine,
} from "./PairingCeremonyMachine.js";

describe("acceptor state machine", () => {
  it("starts in Idle", () => {
    const m = new PairingCeremonyAcceptorMachine();
    assert.equal(m.state, PairingCeremonyAcceptorState.Idle);
  });

  it("walks the happy path with the spec's action firing order", () => {
    const m = new PairingCeremonyAcceptorMachine();
    const fired: PairingCeremonyProtocol.ActionID[] = [];
    for (const id of [
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.RegisterRelay,
      PairingCeremonyProtocol.ActionID.EmitToken,
      PairingCeremonyProtocol.ActionID.StoreCommit,
      PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive,
      PairingCeremonyProtocol.ActionID.StoreRecord,
    ]) {
      m.actions.set(id, () => { fired.push(id); });
    }

    m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin);
    assert.equal(m.state, PairingCeremonyAcceptorState.GeneratingEphemeral);
    m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady);
    assert.equal(m.state, PairingCeremonyAcceptorState.RegisteringRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    assert.equal(m.state, PairingCeremonyAcceptorState.WaitingForHello);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello);
    assert.equal(m.state, PairingCeremonyAcceptorState.WaitingForReveal);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvReveal);
    assert.equal(m.state, PairingCeremonyAcceptorState.DerivingCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady);
    assert.equal(m.state, PairingCeremonyAcceptorState.AwaitingUserConfirm);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserConfirm);
    assert.equal(m.state, PairingCeremonyAcceptorState.AwaitingPeerConfirm);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvConfirmToAcceptor);
    assert.equal(m.state, PairingCeremonyAcceptorState.Paired);

    assert.deepEqual(fired, [
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.RegisterRelay,
      PairingCeremonyProtocol.ActionID.EmitToken,
      PairingCeremonyProtocol.ActionID.StoreCommit,
      PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive,
      PairingCeremonyProtocol.ActionID.StoreRecord,
    ]);
  });

  it("aborts on user_cancel from AwaitingUserConfirm", () => {
    const m = new PairingCeremonyAcceptorMachine();
    for (const id of [
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.RegisterRelay,
      PairingCeremonyProtocol.ActionID.EmitToken,
      PairingCeremonyProtocol.ActionID.StoreCommit,
      PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive,
    ]) {
      m.actions.set(id, () => {});
    }
    m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin);
    m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvReveal);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserCancel);
    assert.equal(m.state, PairingCeremonyAcceptorState.Aborted);
  });

  it("aborts on user_cancel from AwaitingPeerConfirm", () => {
    const m = new PairingCeremonyAcceptorMachine();
    for (const id of [
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.RegisterRelay,
      PairingCeremonyProtocol.ActionID.EmitToken,
      PairingCeremonyProtocol.ActionID.StoreCommit,
      PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive,
    ]) {
      m.actions.set(id, () => {});
    }
    m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin);
    m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvReveal);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserConfirm);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserCancel);
    assert.equal(m.state, PairingCeremonyAcceptorState.Aborted);
  });

  it("aborts on commit_fail from WaitingForReveal", () => {
    const m = new PairingCeremonyAcceptorMachine();
    for (const id of [
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.RegisterRelay,
      PairingCeremonyProtocol.ActionID.EmitToken,
      PairingCeremonyProtocol.ActionID.StoreCommit,
    ]) {
      m.actions.set(id, () => {});
    }
    m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin);
    m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello);
    assert.equal(m.state, PairingCeremonyAcceptorState.WaitingForReveal);
    m.handleEvent(PairingCeremonyProtocol.EventID.CommitFail);
    assert.equal(m.state, PairingCeremonyAcceptorState.Aborted);
  });

  it("propagates an action that throws", () => {
    const m = new PairingCeremonyAcceptorMachine();
    class Boom extends Error {}
    m.actions.set(PairingCeremonyProtocol.ActionID.GenEphemeral, () => {
      throw new Boom();
    });
    assert.throws(() => m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin), Boom);
  });
});

describe("initiator state machine", () => {
  it("starts in Idle", () => {
    const m = new PairingCeremonyInitiatorMachine();
    assert.equal(m.state, PairingCeremonyInitiatorState.Idle);
  });

  it("walks the happy path with the spec's action firing order", () => {
    const m = new PairingCeremonyInitiatorMachine();
    const fired: PairingCeremonyProtocol.ActionID[] = [];
    for (const id of [
      PairingCeremonyProtocol.ActionID.DecodeToken,
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.DialRelay,
      PairingCeremonyProtocol.ActionID.SendReveal,
      PairingCeremonyProtocol.ActionID.DeriveCode,
      PairingCeremonyProtocol.ActionID.StoreRecord,
    ]) {
      m.actions.set(id, () => { fired.push(id); });
    }

    m.handleEvent(PairingCeremonyProtocol.EventID.TokenReceived);
    assert.equal(m.state, PairingCeremonyInitiatorState.DecodingToken);
    m.handleEvent(PairingCeremonyProtocol.EventID.TokenDecoded);
    assert.equal(m.state, PairingCeremonyInitiatorState.GeneratingEphemeral);
    m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady);
    assert.equal(m.state, PairingCeremonyInitiatorState.ConnectingRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);
    assert.equal(m.state, PairingCeremonyInitiatorState.AwaitingWelcome);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvWelcome);
    assert.equal(m.state, PairingCeremonyInitiatorState.Revealing);
    m.handleEvent(PairingCeremonyProtocol.EventID.RevealSent);
    assert.equal(m.state, PairingCeremonyInitiatorState.DerivingCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady);
    assert.equal(m.state, PairingCeremonyInitiatorState.AwaitingUserConfirm);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserConfirm);
    assert.equal(m.state, PairingCeremonyInitiatorState.AwaitingPeerConfirm);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvConfirmToInitiator);
    assert.equal(m.state, PairingCeremonyInitiatorState.Paired);

    assert.deepEqual(fired, [
      PairingCeremonyProtocol.ActionID.DecodeToken,
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.DialRelay,
      PairingCeremonyProtocol.ActionID.SendReveal,
      PairingCeremonyProtocol.ActionID.DeriveCode,
      PairingCeremonyProtocol.ActionID.StoreRecord,
    ]);
  });

  it("aborts on user_cancel", () => {
    const m = new PairingCeremonyInitiatorMachine();
    for (const id of [
      PairingCeremonyProtocol.ActionID.DecodeToken,
      PairingCeremonyProtocol.ActionID.GenEphemeral,
      PairingCeremonyProtocol.ActionID.DialRelay,
      PairingCeremonyProtocol.ActionID.SendReveal,
      PairingCeremonyProtocol.ActionID.DeriveCode,
    ]) {
      m.actions.set(id, () => {});
    }
    m.handleEvent(PairingCeremonyProtocol.EventID.TokenReceived);
    m.handleEvent(PairingCeremonyProtocol.EventID.TokenDecoded);
    m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvWelcome);
    m.handleEvent(PairingCeremonyProtocol.EventID.RevealSent);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserCancel);
    assert.equal(m.state, PairingCeremonyInitiatorState.Aborted);
  });
});

describe("protocol surface", () => {
  it("MessageType wire values match the YAML", () => {
    assert.equal(PairingCeremonyProtocol.MessageType.Hello, "hello");
    assert.equal(PairingCeremonyProtocol.MessageType.Welcome, "welcome");
    assert.equal(PairingCeremonyProtocol.MessageType.Reveal, "reveal");
    assert.equal(PairingCeremonyProtocol.MessageType.ConfirmToInitiator, "confirm_to_initiator");
    assert.equal(PairingCeremonyProtocol.MessageType.ConfirmToAcceptor, "confirm_to_acceptor");
  });

  it("ActionID wire values match the YAML", () => {
    assert.equal(PairingCeremonyProtocol.ActionID.GenEphemeral, "gen_ephemeral");
    assert.equal(PairingCeremonyProtocol.ActionID.RegisterRelay, "register_relay");
    assert.equal(PairingCeremonyProtocol.ActionID.EmitToken, "emit_token");
    assert.equal(PairingCeremonyProtocol.ActionID.StoreCommit, "store_commit");
    assert.equal(PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive, "verify_commit_and_derive");
    assert.equal(PairingCeremonyProtocol.ActionID.StoreRecord, "store_record");
    assert.equal(PairingCeremonyProtocol.ActionID.DecodeToken, "decode_token");
    assert.equal(PairingCeremonyProtocol.ActionID.DialRelay, "dial_relay");
    assert.equal(PairingCeremonyProtocol.ActionID.SendReveal, "send_reveal");
    assert.equal(PairingCeremonyProtocol.ActionID.DeriveCode, "derive_code");
  });
});
