// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  PairingCeremonyServerState,
  PairingCeremonyIosState,
  PairingCeremonyCliState,
  PairingCeremonyProtocol,
  PairingCeremonyServerMachine,
  PairingCeremonyIosMachine,
  PairingCeremonyCliMachine,
} from "./PairingCeremonyMachine.js";

// Helper: walk a PairingCeremonyServerMachine to a target state via the happy path.
function serverAtState(target: PairingCeremonyServerState): PairingCeremonyServerMachine {
  const m = new PairingCeremonyServerMachine();
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
    [PairingCeremonyServerState.Idle]: [],
    [PairingCeremonyServerState.GenerateToken]: [PairingCeremonyProtocol.EventID.RecvPairBegin],
    [PairingCeremonyServerState.RegisterRelay]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated],
    [PairingCeremonyServerState.WaitingForClient]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered],
    [PairingCeremonyServerState.DeriveSecret]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello],
    [PairingCeremonyServerState.SendAck]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete],
    [PairingCeremonyServerState.WaitingForCode]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay],
    [PairingCeremonyServerState.ValidateCode]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit],
    [PairingCeremonyServerState.StorePaired]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode],
    [PairingCeremonyServerState.Paired]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise],
    [PairingCeremonyServerState.AuthCheck]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest],
    [PairingCeremonyServerState.SessionActive]: [PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest, PairingCeremonyProtocol.EventID.Verify],
  };
  for (const ev of paths[target]) {
    m.handleEvent(ev);
  }
  return m;
}

describe("PairingCeremonyServerMachine", () => {
  it("starts in Idle", () => {
    const m = new PairingCeremonyServerMachine();
    assert.equal(m.state, PairingCeremonyServerState.Idle);
  });

  it("Idle -> GenerateToken on recv pair_begin", () => {
    const m = new PairingCeremonyServerMachine();
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.GenerateToken, () => { actionCalled = true; });

    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin);
    assert.equal(m.state, PairingCeremonyServerState.GenerateToken);
    assert.equal(actionCalled, true);
    assert.deepEqual(cmds, []);
    assert.equal(m.currentToken, "tok_1");
  });

  it("GenerateToken -> RegisterRelay on token created", () => {
    const m = serverAtState(PairingCeremonyServerState.GenerateToken);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.RegisterRelay, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated);
    assert.equal(m.state, PairingCeremonyServerState.RegisterRelay);
    assert.equal(actionCalled, true);
  });

  it("RegisterRelay -> WaitingForClient on relay registered", () => {
    const m = serverAtState(PairingCeremonyServerState.RegisterRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    assert.equal(m.state, PairingCeremonyServerState.WaitingForClient);
  });

  it("token_valid guard allows transition to DeriveSecret", () => {
    const m = serverAtState(PairingCeremonyServerState.WaitingForClient);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => false);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello);
    assert.equal(m.state, PairingCeremonyServerState.DeriveSecret);
    assert.equal(actionCalled, true);
    assert.equal(m.serverEcdhPub, "server_pub");
  });

  it("token_invalid guard resets to Idle", () => {
    const m = serverAtState(PairingCeremonyServerState.WaitingForClient);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => true);

    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello);
    assert.equal(m.state, PairingCeremonyServerState.Idle);
  });

  it("code_correct guard transitions to StorePaired", () => {
    const m = serverAtState(PairingCeremonyServerState.ValidateCode);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => false);

    m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode);
    assert.equal(m.state, PairingCeremonyServerState.StorePaired);
  });

  it("code_wrong guard resets to Idle", () => {
    const m = serverAtState(PairingCeremonyServerState.ValidateCode);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => true);

    m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode);
    assert.equal(m.state, PairingCeremonyServerState.Idle);
  });

  it("finalise stores device and transitions to Paired", () => {
    const m = serverAtState(PairingCeremonyServerState.StorePaired);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreDevice, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.Finalise);
    assert.equal(m.state, PairingCeremonyServerState.Paired);
    assert.equal(actionCalled, true);
    assert.equal(m.deviceSecret, "dev_secret_1");
  });

  it("device_known guard transitions to SessionActive", () => {
    const m = serverAtState(PairingCeremonyServerState.AuthCheck);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceKnown, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceUnknown, () => false);
    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.VerifyDevice, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.Verify);
    assert.equal(m.state, PairingCeremonyServerState.SessionActive);
    assert.equal(actionCalled, true);
  });

  it("device_unknown guard resets to Idle", () => {
    const m = serverAtState(PairingCeremonyServerState.AuthCheck);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceKnown, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.DeviceUnknown, () => true);

    m.handleEvent(PairingCeremonyProtocol.EventID.Verify);
    assert.equal(m.state, PairingCeremonyServerState.Idle);
  });

  it("disconnect returns to Paired", () => {
    const m = serverAtState(PairingCeremonyServerState.SessionActive);
    m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, PairingCeremonyServerState.Paired);
  });

  it("invalid event does not change state", () => {
    const m = new PairingCeremonyServerMachine();
    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, PairingCeremonyServerState.Idle);
    assert.deepEqual(cmds, []);
  });

  it("full pairing flow", () => {
    const m = new PairingCeremonyServerMachine();
    m.actions.set(PairingCeremonyProtocol.ActionID.GenerateToken, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.RegisterRelay, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreDevice, () => {});
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenValid, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.TokenInvalid, () => false);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeCorrect, () => true);
    m.guards.set(PairingCeremonyProtocol.GuardID.CodeWrong, () => false);

    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin);
    assert.equal(m.state, PairingCeremonyServerState.GenerateToken);
    m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated);
    assert.equal(m.state, PairingCeremonyServerState.RegisterRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered);
    assert.equal(m.state, PairingCeremonyServerState.WaitingForClient);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello);
    assert.equal(m.state, PairingCeremonyServerState.DeriveSecret);
    m.handleEvent(PairingCeremonyProtocol.EventID.ECDHComplete);
    assert.equal(m.state, PairingCeremonyServerState.SendAck);
    m.handleEvent(PairingCeremonyProtocol.EventID.SignalCodeDisplay);
    assert.equal(m.state, PairingCeremonyServerState.WaitingForCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvCodeSubmit);
    assert.equal(m.state, PairingCeremonyServerState.ValidateCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode);
    assert.equal(m.state, PairingCeremonyServerState.StorePaired);
    m.handleEvent(PairingCeremonyProtocol.EventID.Finalise);
    assert.equal(m.state, PairingCeremonyServerState.Paired);
  });
});

describe("PairingCeremonyIosMachine", () => {
  it("starts in Idle", () => {
    const m = new PairingCeremonyIosMachine();
    assert.equal(m.state, PairingCeremonyIosState.Idle);
  });

  it("full pairing flow", () => {
    const m = new PairingCeremonyIosMachine();
    m.actions.set(PairingCeremonyProtocol.ActionID.SendPairHello, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.DeriveSecret, () => {});
    m.actions.set(PairingCeremonyProtocol.ActionID.StoreSecret, () => {});

    m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR);
    assert.equal(m.state, PairingCeremonyIosState.ScanQR);
    m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed);
    assert.equal(m.state, PairingCeremonyIosState.ConnectRelay);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);
    assert.equal(m.state, PairingCeremonyIosState.GenKeyPair);
    m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated);
    assert.equal(m.state, PairingCeremonyIosState.WaitAck);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHelloAck);
    assert.equal(m.state, PairingCeremonyIosState.E2EReady);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairConfirm);
    assert.equal(m.state, PairingCeremonyIosState.ShowCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.CodeDisplayed);
    assert.equal(m.state, PairingCeremonyIosState.WaitPairComplete);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairComplete);
    assert.equal(m.state, PairingCeremonyIosState.Paired);
  });

  it("key pair generated calls sendPairHello action", () => {
    const m = new PairingCeremonyIosMachine();
    m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR);
    m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);

    let actionCalled = false;
    m.actions.set(PairingCeremonyProtocol.ActionID.SendPairHello, () => { actionCalled = true; });

    m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated);
    assert.equal(actionCalled, true);
    assert.equal(m.state, PairingCeremonyIosState.WaitAck);
  });

  it("reconnect and auth flow", () => {
    const m = new PairingCeremonyIosMachine();
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
    assert.equal(m.state, PairingCeremonyIosState.Paired);

    m.handleEvent(PairingCeremonyProtocol.EventID.AppLaunch);
    assert.equal(m.state, PairingCeremonyIosState.Reconnect);
    m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected);
    assert.equal(m.state, PairingCeremonyIosState.SendAuth);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthOk);
    assert.equal(m.state, PairingCeremonyIosState.SessionActive);
    m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, PairingCeremonyIosState.Paired);
  });

  it("invalid event does not change state", () => {
    const m = new PairingCeremonyIosMachine();
    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect);
    assert.equal(m.state, PairingCeremonyIosState.Idle);
    assert.deepEqual(cmds, []);
  });
});

describe("PairingCeremonyCliMachine", () => {
  it("starts in Idle", () => {
    const m = new PairingCeremonyCliMachine();
    assert.equal(m.state, PairingCeremonyCliState.Idle);
  });

  it("full flow", () => {
    const m = new PairingCeremonyCliMachine();

    m.handleEvent(PairingCeremonyProtocol.EventID.CliInit);
    assert.equal(m.state, PairingCeremonyCliState.GetKey);
    m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored);
    assert.equal(m.state, PairingCeremonyCliState.BeginPair);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvTokenResponse);
    assert.equal(m.state, PairingCeremonyCliState.ShowQR);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvWaitingForCode);
    assert.equal(m.state, PairingCeremonyCliState.PromptCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.UserEntersCode);
    assert.equal(m.state, PairingCeremonyCliState.SubmitCode);
    m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairStatus);
    assert.equal(m.state, PairingCeremonyCliState.Done);
  });

  it("invalid event does not change state", () => {
    const m = new PairingCeremonyCliMachine();
    const cmds = m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored);
    assert.equal(m.state, PairingCeremonyCliState.Idle);
    assert.deepEqual(cmds, []);
  });
});
