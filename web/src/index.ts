// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

export {
  E2EKeyPair,
  E2EChannel,
  deriveKeyFromSecret,
  deriveConfirmationCode,
  generateNonce,
  generateSecret,
  createPairingRecord,
  deriveChannelFromRecord,
  type PairingRecord,
} from "./crypto.js";
export {
  connect,
  wakeRelay,
  Session,
  Stream,
  Datagram,
  encodeUvarint,
  decodeUvarint,
  encodeStreamHeader,
  decodeStreamHeader,
  encodeDatagram,
  decodeDatagram,
  type ConnectArgs,
  type Identity,
} from "./relay.js";
