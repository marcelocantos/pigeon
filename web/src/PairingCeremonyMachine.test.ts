// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  ServerState,
  IosState,
  CliState,
  PairingCeremonyProtocol,
  ServerMachine,
  IosMachine,
  CliMachine,
} from "./PairingCeremonyMachine.js";

// Helper: walk a ServerMachine to a target state via the happy path.
function serverAtState(target: ServerState): ServerMachine {
  const m = new ServerMachine();
  m.actions.set(PairingCeremonyProtocol.ActionID.GenerateToken, () => {});
  m.actions.set(PairingCeremonyProtocol.ActionID.RegisterRelay, () => {});
  m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => {});
  m.actions.set(PairingCeremonyProtocol.ActionID.StoreDevice, () => {});
  m.actions.set(PairingCeremonyProtocol.ActionID.VerifyDevice, () => {});
  m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => true);
  m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => false);
  m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => true);
  m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => false);
  m.guards.set(PairingCeremonyProtocol.GuardID.DeviceKnown, () => true);
  m.guards.set(PairingCeremonyProtocol.GuardID.DeviceUnknown, () => false);

  const paths: Record<string, PairingCeremonyProtocol.EventID[]> = {
    [ServerState.Idle]: [],
    [ServerState.GenerateToken]: [PairingCeremonyProtocol.EventID.RecvPairBegin],
    [ServerState.RegisterRelay]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated],
    [ServerState.WaitingForClient]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered],
    [ServerState.DeriveSecret]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello],
    [ServerState.SendAck]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete],
    [ServerState.WaitingForCode]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay],
    [ServerState.ValidateCode]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit],
    [ServerState.StorePaired]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode],
    [ServerState.Paired]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise],
    [ServerState.AuthCheck]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest],
    [ServerState.SessionActive]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest, PairingCeremonyProtocol.EventID.Verify],
  };
  for (const ev of paths[target]) {
    m.handleEvent(ev);
  }
  return m;
}

describe("ServerMachine", () => {
  it("starts in Idle", () => {
    const m = new ServerMachine();
    assert.equal(m.state, ServerState.Idle);
  });

  it("Idle -> GenerateToken on recv pair_begin", () => {
    const m = new ServerMachine();
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.GenerateToken, () => { actionCalled = true; });

    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin);
    assert.equal(m.state, ServerState.GenerateToken);
    assert.equal(actionCalled, true);
    assert.deepEqual(cmds, []);
    assert.equal(m.currentToken, "tok_1");
  });

  it("GenerateToken -> RegisterRelay on token created", () => {
    const m = serverAtState(ServerState.GenerateToken);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.RegisterRelay, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated);
    assert.equal(m.state, ServerState.RegisterRelay);
    assert.equal(actionCalled, true);
  });

  it("RegisterRelay -> WaitingForClient on relay registered", () => {
    const m = serverAtState(ServerState.RegisterRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    assert.equal(m.state, ServerState.WaitingForClient);
  });

  it("token_valid guard allows transition to DeriveSecret", () => {
    const m = serverAtState(ServerState.WaitingForClient);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => false);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello);
    assert.equal(m.state, ServerState.DeriveSecret);
    assert.equal(actionCalled, true);
    assert.equal(m.serverEcdhPub, "server_pub");
  });

  it("token_invalid guard resets to Idle", () => {
    const m = serverAtState(ServerState.WaitingForClient);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => true);

    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello);
    assert.equal(m.state, ServerState.Idle);
  });

  it("code_correct guard transitions to StorePaired", () => {
    const m = serverAtState(ServerState.ValidateCode);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => false);

    m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode);
    assert.equal(m.state, ServerState.StorePaired);
  });

  it("code_wrong guard resets to Idle", () => {
    const m = serverAtState(ServerState.ValidateCode);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => true);

    m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode);
    assert.equal(m.state, ServerState.Idle);
  });

  it("finalise stores device and transitions to Paired", () => {
    const m = serverAtState(ServerState.StorePaired);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreDevice, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.Finalise);
    assert.equal(m.state, ServerState.Paired);
    assert.equal(actionCalled, true);
    assert.equal(m.deviceSecret, "dev_secret_1");
  });

  it("device_known guard transitions to SessionActive", () => {
    const m = serverAtState(ServerState.AuthCheck);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceKnown, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceUnknown, () => false);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.VerifyDevice, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.Verify);
    assert.equal(m.state, ServerState.SessionActive);
    assert.equal(actionCalled, true);
  });

  it("device_unknown guard resets to Idle", () => {
    const m = serverAtState(ServerState.AuthCheck);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceKnown, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceUnknown, () => true);

    m.handleEvent(PairingCeremonyProtocol.EventID.Verify);
    assert.equal(m.state, ServerState.Idle);
  });

  it("disconnect returns to Paired", () => {
    const m = serverAtState(ServerState.SessionActive);
    m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, ServerState.Paired);
  });

  it("invalid event does not change state", () => {
    const m = new ServerMachine();
    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, ServerState.Idle);
    assert.deepEqual(cmds, []);
  });

  it("full pairing flow", () => {
    const m = new ServerMachine();
    m.actions.set(PairingCeremonyProtocol.ActionID.GenerateToken, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.RegisterRelay, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreDevice, () => {});
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => false);

    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin);
    assert.equal(m.state, ServerState.GenerateToken);
    m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated);
    assert.equal(m.state, ServerState.RegisterRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    assert.equal(m.state, ServerState.WaitingForClient);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello);
    assert.equal(m.state, ServerState.DeriveSecret);
    m.handleEvent(PairingCeremonyProtocol.EventID.ECDHComplete);
    assert.equal(m.state, ServerState.SendAck);
    m.handleEvent(PairingCeremonyProtocol.EventID.SignalCodeDisplay);
    assert.equal(m.state, ServerState.WaitingForCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvCodeSubmit);
    assert.equal(m.state, ServerState.ValidateCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode);
    assert.equal(m.state, ServerState.StorePaired);
    m.handleEvent(PairingCeremonyProtocol.EventID.Finalise);
    assert.equal(m.state, ServerState.Paired);
  });
});

describe("IosMachine", () => {
  it("starts in Idle", () => {
    const m = new IosMachine();
    assert.equal(m.state, IosState.Idle);
  });

  it("full pairing flow", () => {
    const m = new IosMachine();
    m.actions.set(PairingCeremonyProtocol.ActionID.SendPairHello, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreSecret, () => {});

    m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR);
    assert.equal(m.state, IosState.ScanQR);
    m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed);
    assert.equal(m.state, IosState.ConnectRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);
    assert.equal(m.state, IosState.GenKeyPair);
    m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated);
    assert.equal(m.state, IosState.WaitAck);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHelloAck);
    assert.equal(m.state, IosState.E2EReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairConfirm);
    assert.equal(m.state, IosState.ShowCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeDisplayed);
    assert.equal(m.state, IosState.WaitPairComplete);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairComplete);
    assert.equal(m.state, IosState.Paired);
  });

  it("key pair generated calls sendPairHello action", () => {
    const m = new IosMachine();
    m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR);
    m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);

    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.SendPairHello, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated);
    assert.equal(actionCalled, true);
    assert.equal(m.state, IosState.WaitAck);
  });

  it("reconnect and auth flow", () => {
    const m = new IosMachine();
    m.actions.set(PairingCeremonyProtocol.ActionID.SendPairHello, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreSecret, () => {});

    // Walk to Paired
    for (const ev of [
      PairingCeremonyProtocol.EventID.UserScansQR,
      PairingCeremonyProtocol.EventID.QRParsed,
      PairingCeremonyProtocol.EventID.RelayConnected,
      PairingCeremonyProtocol.EventID.KeyPairGenerated,
      PairingCeremonyProtocol.EventID.RecvPairHelloAck,
      PairingCeremonyProtocol.EventID.RecvPairConfirm,
      PairingCeremonyProtocol.EventID.CodeDisplayed,
      PairingCeremonyProtocol.EventID.RecvPairComplete,
    ]) {
      m.handleEvent(ev);
    }
    assert.equal(m.state, IosState.Paired);

    m.handleEvent(PairingCeremonyProtocol.EventID.AppLaunch);
    assert.equal(m.state, IosState.Reconnect);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);
    assert.equal(m.state, IosState.SendAuth);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthOk);
    assert.equal(m.state, IosState.SessionActive);
    m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, IosState.Paired);
  });

  it("invalid event does not change state", () => {
    const m = new IosMachine();
    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, IosState.Idle);
    assert.deepEqual(cmds, []);
  });
});

describe("CliMachine", () => {
  it("starts in Idle", () => {
    const m = new CliMachine();
    assert.equal(m.state, CliState.Idle);
  });

  it("full flow", () => {
    const m = new CliMachine();

    m.handleEvent(PairingCeremonyProtocol.EventID.CliInit);
    assert.equal(m.state, CliState.GetKey);
    m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored);
    assert.equal(m.state, CliState.BeginPair);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvTokenResponse);
    assert.equal(m.state, CliState.ShowQR);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvWaitingForCode);
    assert.equal(m.state, CliState.PromptCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserEntersCode);
    assert.equal(m.state, CliState.SubmitCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairStatus);
    assert.equal(m.state, CliState.Done);
  });

  it("invalid event does not change state", () => {
    const m = new CliMachine();
    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored);
    assert.equal(m.state, CliState.Idle);
    assert.deepEqual(cmds, []);
  });
});
