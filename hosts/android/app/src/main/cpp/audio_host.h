#pragma once
#include "pcm_ring.h"
// Lazily constructed process-owned output; lifetime exceeds all Oboe callbacks.
// Request methods only change atomics. Dedicated control thread owns streams.
a4::PcmBuffer& androidAudioBuffer();
