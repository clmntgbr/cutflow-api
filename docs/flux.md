┌──────────────────────────────────────────────┐
│ 1. UPLOAD                                    │
│                                              │
│ Upload original.mp4 → S3 / MinIO             │
│                                              │
│ Création :                                   │
│ • Project                                    │
│ • MediaFile                                  │
│ • MediaConfiguration avec valeurs par défaut │
└──────────────────────┬───────────────────────┘
                       ▼
┌──────────────────────────────────────────────┐
│ 2. MEDIA PROBE                               │
│                                              │
│ FFprobe → duration, codecs, …                │
└──────────────────────┬───────────────────────┘
                       ▼
┌──────────────────────────────────────────────┐
│ 3. EXTRACTION AUDIO                          │
│                                              │
│ original.mp4 → audio.opus                    │
└──────────────────────┬───────────────────────┘
                       │
              ┌────────┴────────┐
              ▼                 ▼
┌─────────────────────────────┐    ┌─────────────────────────────┐
│ 4A. TRANSCRIPTION           │    │ 4B. AUDIO ANALYSIS          │
│ AssemblyAI                  │    │ DetectedSilence[] (brut)    │
│ TranscriptWord[] + SRT      │    │ + noise floor / threshold   │
└──────────────┬──────────────┘    └──────────────┬──────────────┘
               │                                  │
               │ transcript.completed             │
      ┌────────┴────────┐                         │
      ▼                 ▼                         │
┌───────────────┐  ┌────────────────┐             │
│ 5A. TEXT      │  │ 5B. VIRAL      │             │
│ ANALYSIS      │  │ ANALYSIS       │             │
│ fillers /     │  │ ViralCandidate │             │
│ répétitions / │  │ (SOURCE TIME)  │             │
│ faux départs  │  │ (non bloquant) │             │
└───────┬───────┘  └────────────────┘             │
        │                                         │
        │              ┌──────────────────────────┘
        │              │
        ▼              ▼
┌──────────────────────────────────────────────┐
│ BARRIER (initial only)                       │
│                                              │
│ wait silence ✓ AND text-analysis ✓           │
│ viral n'est PAS dans la barrière             │
│                                              │
│ → timeline_rebuild_requested                 │
│   reason=initial_analysis_ready              │
└──────────────────────┬───────────────────────┘
                       ▼
┌──────────────────────────────────────────────┐
│ 6. TIMELINE WORKER → Timeline V1             │
│                                              │
│ DetectedSilence[]                            │
│ DetectedTranscriptIssue[]                    │
│ MediaConfiguration                           │
│ UserOverride[]                               │
│       ↓                                      │
│ EditDecision[] → Timeline                    │
│       ↓                                      │
│ READY_TO_EDIT                                │
└──────────────────────┬───────────────────────┘
                       ▼
┌──────────────────────────────────────────────┐
│ AFTER INITIAL: rebuild immédiat              │
│                                              │
│ config / override / re-analyse               │
│ → V2, V3, V4…                                │
└──────────────────────────────────────────────┘
