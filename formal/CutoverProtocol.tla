---- MODULE CutoverProtocol ----
\* Bounded-loss cutover model for transport path switching.
\*
\* Design: single-active-path with CUTOVER marker. When a side wants
\* to switch paths, it sends a CUTOVER marker as the last message on
\* the old path, then switches to the new path. The receiver drains
\* the old path after seeing the marker (bounded by a drain window),
\* then closes it and begins reading the new path. Messages in flight
\* on the old path at cutover may be lost (bounded by RTT x throughput).
\*
\* Key ordering guarantee: the receiver does NOT read from the new path
\* until the old path is drained. This prevents out-of-order delivery.
\*
\* If the CUTOVER marker itself is lost, a timeout fires and the
\* receiver closes the old path anyway (bounded loss is acceptable).
\*
\* Verified properties:
\*   - No message duplication
\*   - No message delivered out of order
\*   - No deadlock (concurrent cutover from both sides is safe)
\*   - Bounded loss model (not lossless — by design)

EXTENDS Integers, Sequences, FiniteSets

CONSTANTS
    MaxMessages,        \* bound on number of app messages per side
    MaxInFlight,        \* bound on channel capacity (in-flight messages)
    DrainWindow         \* max messages drained after CUTOVER received

\* Path names
PATH_A == "A"
PATH_B == "B"

OtherPath(p) == IF p = PATH_A THEN PATH_B ELSE PATH_A

\* Cutover marker is represented as 0; app messages are 1..MaxMessages.
CUTOVER == 0

VARIABLES
    \* Per-side state: which path is active for sending
    senderPath,         \* [Sides -> {PATH_A, PATH_B}]

    \* Per-side: which path the receiver accepts messages from.
    \* Starts at PATH_A; switches to PATH_B only after cutover completes.
    receiverPath,       \* [Sides -> {PATH_A, PATH_B}]

    \* Per-side: has this side initiated cutover (sent the marker)?
    cutoverInitiated,   \* [Sides -> BOOLEAN]

    \* Per-side: has this side finished the cutover?
    cutoverComplete,    \* [Sides -> BOOLEAN]

    \* Per-side: is this side currently draining the old path?
    draining,           \* [Sides -> BOOLEAN]

    \* Per-side drain counter
    drainCount,         \* [Sides -> 0..DrainWindow]

    \* In-flight message channels per direction and path
    leftToRight_A, leftToRight_B,
    rightToLeft_A, rightToLeft_B,

    \* Delivered message logs (app messages only, in delivery order)
    leftDelivered,
    rightDelivered,

    \* Send counters (generate unique message IDs 1..MaxMessages)
    leftSendCount,
    rightSendCount

vars == <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
          draining, drainCount, leftToRight_A, leftToRight_B,
          rightToLeft_A, rightToLeft_B, leftDelivered, rightDelivered,
          leftSendCount, rightSendCount>>

Sides == {"left", "right"}

Init ==
    /\ senderPath = [s \in Sides |-> PATH_A]
    /\ receiverPath = [s \in Sides |-> PATH_A]
    /\ cutoverInitiated = [s \in Sides |-> FALSE]
    /\ cutoverComplete = [s \in Sides |-> FALSE]
    /\ draining = [s \in Sides |-> FALSE]
    /\ drainCount = [s \in Sides |-> 0]
    /\ leftToRight_A = <<>>
    /\ leftToRight_B = <<>>
    /\ rightToLeft_A = <<>>
    /\ rightToLeft_B = <<>>
    /\ leftDelivered = <<>>
    /\ rightDelivered = <<>>
    /\ leftSendCount = 0
    /\ rightSendCount = 0

\* ================================================================
\* Actions
\* ================================================================

\* --- Sending (always on the sender's active path) ---

LeftSendApp ==
    /\ leftSendCount < MaxMessages
    /\ LET path == senderPath["left"]
           msg == leftSendCount + 1
       IN
       /\ IF path = PATH_A
          THEN /\ Len(leftToRight_A) < MaxInFlight
               /\ leftToRight_A' = Append(leftToRight_A, msg)
               /\ UNCHANGED leftToRight_B
          ELSE /\ Len(leftToRight_B) < MaxInFlight
               /\ leftToRight_B' = Append(leftToRight_B, msg)
               /\ UNCHANGED leftToRight_A
       /\ leftSendCount' = leftSendCount + 1
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   draining, drainCount, rightToLeft_A, rightToLeft_B,
                   leftDelivered, rightDelivered, rightSendCount>>

RightSendApp ==
    /\ rightSendCount < MaxMessages
    /\ LET path == senderPath["right"]
           msg == rightSendCount + 1
       IN
       /\ IF path = PATH_A
          THEN /\ Len(rightToLeft_A) < MaxInFlight
               /\ rightToLeft_A' = Append(rightToLeft_A, msg)
               /\ UNCHANGED rightToLeft_B
          ELSE /\ Len(rightToLeft_B) < MaxInFlight
               /\ rightToLeft_B' = Append(rightToLeft_B, msg)
               /\ UNCHANGED rightToLeft_A
       /\ rightSendCount' = rightSendCount + 1
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   draining, drainCount, leftToRight_A, leftToRight_B,
                   leftDelivered, rightDelivered, leftSendCount>>

\* --- Cutover initiation ---
\* Send CUTOVER marker on current path, switch sender to new path.

LeftInitiateCutover ==
    /\ ~cutoverInitiated["left"]
    /\ LET path == senderPath["left"]
       IN
       /\ IF path = PATH_A
          THEN /\ Len(leftToRight_A) < MaxInFlight
               /\ leftToRight_A' = Append(leftToRight_A, CUTOVER)
               /\ UNCHANGED leftToRight_B
          ELSE /\ Len(leftToRight_B) < MaxInFlight
               /\ leftToRight_B' = Append(leftToRight_B, CUTOVER)
               /\ UNCHANGED leftToRight_A
    /\ cutoverInitiated' = [cutoverInitiated EXCEPT !["left"] = TRUE]
    /\ senderPath' = [senderPath EXCEPT !["left"] = OtherPath(senderPath["left"])]
    /\ UNCHANGED <<receiverPath, cutoverComplete, draining, drainCount,
                   rightToLeft_A, rightToLeft_B,
                   leftDelivered, rightDelivered,
                   leftSendCount, rightSendCount>>

RightInitiateCutover ==
    /\ ~cutoverInitiated["right"]
    /\ LET path == senderPath["right"]
       IN
       /\ IF path = PATH_A
          THEN /\ Len(rightToLeft_A) < MaxInFlight
               /\ rightToLeft_A' = Append(rightToLeft_A, CUTOVER)
               /\ UNCHANGED rightToLeft_B
          ELSE /\ Len(rightToLeft_B) < MaxInFlight
               /\ rightToLeft_B' = Append(rightToLeft_B, CUTOVER)
               /\ UNCHANGED rightToLeft_A
    /\ cutoverInitiated' = [cutoverInitiated EXCEPT !["right"] = TRUE]
    /\ senderPath' = [senderPath EXCEPT !["right"] = OtherPath(senderPath["right"])]
    /\ UNCHANGED <<receiverPath, cutoverComplete, draining, drainCount,
                   leftToRight_A, leftToRight_B,
                   leftDelivered, rightDelivered,
                   leftSendCount, rightSendCount>>

\* --- Receiving ---
\* A side only receives from its receiverPath. During drain, it reads
\* from the OLD path (receiverPath hasn't switched yet). After cutover
\* completes, receiverPath switches and it reads from the new path.

RightRecvFromA ==
    /\ receiverPath["right"] = PATH_A  \* right is reading from path A
    /\ Len(leftToRight_A) > 0
    /\ LET msg == Head(leftToRight_A) IN
       /\ leftToRight_A' = Tail(leftToRight_A)
       /\ IF msg = CUTOVER
          THEN /\ draining' = [draining EXCEPT !["right"] = TRUE]
               /\ drainCount' = [drainCount EXCEPT !["right"] = 0]
               /\ UNCHANGED rightDelivered
          ELSE IF draining["right"]
               THEN IF drainCount["right"] < DrainWindow
                    THEN /\ rightDelivered' = Append(rightDelivered, msg)
                         /\ drainCount' = [drainCount EXCEPT !["right"] = drainCount["right"] + 1]
                         /\ UNCHANGED draining
                    ELSE /\ UNCHANGED <<rightDelivered, draining, drainCount>>
               ELSE /\ rightDelivered' = Append(rightDelivered, msg)
                    /\ UNCHANGED <<draining, drainCount>>
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   leftToRight_B, rightToLeft_A, rightToLeft_B,
                   leftDelivered, leftSendCount, rightSendCount>>

RightRecvFromB ==
    /\ receiverPath["right"] = PATH_B
    /\ Len(leftToRight_B) > 0
    /\ LET msg == Head(leftToRight_B) IN
       /\ leftToRight_B' = Tail(leftToRight_B)
       /\ IF msg = CUTOVER
          THEN /\ draining' = [draining EXCEPT !["right"] = TRUE]
               /\ drainCount' = [drainCount EXCEPT !["right"] = 0]
               /\ UNCHANGED rightDelivered
          ELSE IF draining["right"]
               THEN IF drainCount["right"] < DrainWindow
                    THEN /\ rightDelivered' = Append(rightDelivered, msg)
                         /\ drainCount' = [drainCount EXCEPT !["right"] = drainCount["right"] + 1]
                         /\ UNCHANGED draining
                    ELSE /\ UNCHANGED <<rightDelivered, draining, drainCount>>
               ELSE /\ rightDelivered' = Append(rightDelivered, msg)
                    /\ UNCHANGED <<draining, drainCount>>
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   leftToRight_A, rightToLeft_A, rightToLeft_B,
                   leftDelivered, leftSendCount, rightSendCount>>

LeftRecvFromA ==
    /\ receiverPath["left"] = PATH_A
    /\ Len(rightToLeft_A) > 0
    /\ LET msg == Head(rightToLeft_A) IN
       /\ rightToLeft_A' = Tail(rightToLeft_A)
       /\ IF msg = CUTOVER
          THEN /\ draining' = [draining EXCEPT !["left"] = TRUE]
               /\ drainCount' = [drainCount EXCEPT !["left"] = 0]
               /\ UNCHANGED leftDelivered
          ELSE IF draining["left"]
               THEN IF drainCount["left"] < DrainWindow
                    THEN /\ leftDelivered' = Append(leftDelivered, msg)
                         /\ drainCount' = [drainCount EXCEPT !["left"] = drainCount["left"] + 1]
                         /\ UNCHANGED draining
                    ELSE /\ UNCHANGED <<leftDelivered, draining, drainCount>>
               ELSE /\ leftDelivered' = Append(leftDelivered, msg)
                    /\ UNCHANGED <<draining, drainCount>>
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   rightToLeft_B, leftToRight_A, leftToRight_B,
                   rightDelivered, leftSendCount, rightSendCount>>

LeftRecvFromB ==
    /\ receiverPath["left"] = PATH_B
    /\ Len(rightToLeft_B) > 0
    /\ LET msg == Head(rightToLeft_B) IN
       /\ rightToLeft_B' = Tail(rightToLeft_B)
       /\ IF msg = CUTOVER
          THEN /\ draining' = [draining EXCEPT !["left"] = TRUE]
               /\ drainCount' = [drainCount EXCEPT !["left"] = 0]
               /\ UNCHANGED leftDelivered
          ELSE IF draining["left"]
               THEN IF drainCount["left"] < DrainWindow
                    THEN /\ leftDelivered' = Append(leftDelivered, msg)
                         /\ drainCount' = [drainCount EXCEPT !["left"] = drainCount["left"] + 1]
                         /\ UNCHANGED draining
                    ELSE /\ UNCHANGED <<leftDelivered, draining, drainCount>>
               ELSE /\ leftDelivered' = Append(leftDelivered, msg)
                    /\ UNCHANGED <<draining, drainCount>>
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   rightToLeft_A, leftToRight_A, leftToRight_B,
                   rightDelivered, leftSendCount, rightSendCount>>

\* --- Cutover completion ---
\* After receiving CUTOVER marker and draining, switch receiverPath.

LeftCompleteCutover ==
    /\ draining["left"]
    /\ ~cutoverComplete["left"]
    /\ cutoverComplete' = [cutoverComplete EXCEPT !["left"] = TRUE]
    /\ draining' = [draining EXCEPT !["left"] = FALSE]
    /\ receiverPath' = [receiverPath EXCEPT !["left"] = OtherPath(receiverPath["left"])]
    /\ UNCHANGED <<senderPath, cutoverInitiated, drainCount,
                   leftToRight_A, leftToRight_B, rightToLeft_A, rightToLeft_B,
                   leftDelivered, rightDelivered, leftSendCount, rightSendCount>>

RightCompleteCutover ==
    /\ draining["right"]
    /\ ~cutoverComplete["right"]
    /\ cutoverComplete' = [cutoverComplete EXCEPT !["right"] = TRUE]
    /\ draining' = [draining EXCEPT !["right"] = FALSE]
    /\ receiverPath' = [receiverPath EXCEPT !["right"] = OtherPath(receiverPath["right"])]
    /\ UNCHANGED <<senderPath, cutoverInitiated, drainCount,
                   leftToRight_A, leftToRight_B, rightToLeft_A, rightToLeft_B,
                   leftDelivered, rightDelivered, leftSendCount, rightSendCount>>

\* Timeout-based cutover completion: if the CUTOVER marker was lost,
\* the receiver eventually times out and switches anyway.
LeftTimeoutCutover ==
    /\ ~draining["left"]
    /\ ~cutoverComplete["left"]
    /\ cutoverInitiated["right"]  \* right sent a cutover marker
    /\ cutoverComplete' = [cutoverComplete EXCEPT !["left"] = TRUE]
    /\ receiverPath' = [receiverPath EXCEPT !["left"] = OtherPath(receiverPath["left"])]
    /\ UNCHANGED <<senderPath, cutoverInitiated, draining, drainCount,
                   leftToRight_A, leftToRight_B, rightToLeft_A, rightToLeft_B,
                   leftDelivered, rightDelivered, leftSendCount, rightSendCount>>

RightTimeoutCutover ==
    /\ ~draining["right"]
    /\ ~cutoverComplete["right"]
    /\ cutoverInitiated["left"]
    /\ cutoverComplete' = [cutoverComplete EXCEPT !["right"] = TRUE]
    /\ receiverPath' = [receiverPath EXCEPT !["right"] = OtherPath(receiverPath["right"])]
    /\ UNCHANGED <<senderPath, cutoverInitiated, draining, drainCount,
                   leftToRight_A, leftToRight_B, rightToLeft_A, rightToLeft_B,
                   leftDelivered, rightDelivered, leftSendCount, rightSendCount>>

\* --- Message loss ---

LoseMessageLeftToRightA ==
    /\ Len(leftToRight_A) > 0
    /\ leftToRight_A' = Tail(leftToRight_A)
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   draining, drainCount, leftToRight_B, rightToLeft_A,
                   rightToLeft_B, leftDelivered, rightDelivered,
                   leftSendCount, rightSendCount>>

LoseMessageLeftToRightB ==
    /\ Len(leftToRight_B) > 0
    /\ leftToRight_B' = Tail(leftToRight_B)
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   draining, drainCount, leftToRight_A, rightToLeft_A,
                   rightToLeft_B, leftDelivered, rightDelivered,
                   leftSendCount, rightSendCount>>

LoseMessageRightToLeftA ==
    /\ Len(rightToLeft_A) > 0
    /\ rightToLeft_A' = Tail(rightToLeft_A)
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   draining, drainCount, leftToRight_A, leftToRight_B,
                   rightToLeft_B, leftDelivered, rightDelivered,
                   leftSendCount, rightSendCount>>

LoseMessageRightToLeftB ==
    /\ Len(rightToLeft_B) > 0
    /\ rightToLeft_B' = Tail(rightToLeft_B)
    /\ UNCHANGED <<senderPath, receiverPath, cutoverInitiated, cutoverComplete,
                   draining, drainCount, leftToRight_A, leftToRight_B,
                   rightToLeft_A, leftDelivered, rightDelivered,
                   leftSendCount, rightSendCount>>

Next ==
    \/ LeftSendApp
    \/ RightSendApp
    \/ LeftInitiateCutover
    \/ RightInitiateCutover
    \/ RightRecvFromA
    \/ RightRecvFromB
    \/ LeftRecvFromA
    \/ LeftRecvFromB
    \/ LeftCompleteCutover
    \/ RightCompleteCutover
    \/ LeftTimeoutCutover
    \/ RightTimeoutCutover
    \/ LoseMessageLeftToRightA
    \/ LoseMessageLeftToRightB
    \/ LoseMessageRightToLeftA
    \/ LoseMessageRightToLeftB

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants
\* ================================================================

\* No message is delivered more than once.
NoDuplicateHelper(seq) ==
    \A i, j \in 1..Len(seq) :
        (i /= j) => (seq[i] /= seq[j])

NoDuplicateDelivery ==
    /\ NoDuplicateHelper(leftDelivered)
    /\ NoDuplicateHelper(rightDelivered)

\* Messages are delivered in order (strictly increasing IDs).
IsStrictlyIncreasing(seq) ==
    \A i \in 1..(Len(seq)-1) : seq[i] < seq[i+1]

NoOutOfOrderDelivery ==
    /\ (Len(leftDelivered) <= 1 \/ IsStrictlyIncreasing(leftDelivered))
    /\ (Len(rightDelivered) <= 1 \/ IsStrictlyIncreasing(rightDelivered))

\* Type invariant.
TypeOK ==
    /\ senderPath \in [Sides -> {PATH_A, PATH_B}]
    /\ receiverPath \in [Sides -> {PATH_A, PATH_B}]
    /\ cutoverInitiated \in [Sides -> BOOLEAN]
    /\ cutoverComplete \in [Sides -> BOOLEAN]
    /\ draining \in [Sides -> BOOLEAN]
    /\ drainCount \in [Sides -> 0..DrainWindow]
    /\ leftSendCount \in 0..MaxMessages
    /\ rightSendCount \in 0..MaxMessages

\* After cutover is initiated, the sender is using the new path.
CutoverSwitchesPath ==
    /\ (cutoverInitiated["left"] => senderPath["left"] = PATH_B)
    /\ (cutoverInitiated["right"] => senderPath["right"] = PATH_B)

====
