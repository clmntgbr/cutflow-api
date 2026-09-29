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
│ FFprobe                                      │
│                                              │
│ Récupération :                               │
│ • durée                                      │
│ • résolution                                 │
│ • framerate                                  │
│ • codec vidéo                                │
│ • codec audio                                │
│ • bitrate                                    │
│ • présence d'une piste audio                 │
│ • taille                                     │
│                                              │
│ + validation du fichier                      │
└──────────────────────┬───────────────────────┘
                       ▼
┌──────────────────────────────────────────────┐
│ 3. EXTRACTION AUDIO                          │
│                                              │
│ FFmpeg                                       │
│                                              │
│ original.mp4                                 │
│      ↓                                       │
│ audio.opus                                   │
│                                              │
│ ~32-40 kbps / mono / 16 kHz                  │
│                                              │
│ Ex : ~1.4 Mo pour 5 minutes                  │
└──────────────────────┬───────────────────────┘
                       │
              ┌────────┴────────┐
              ▼                 ▼
┌─────────────────────────────┐    ┌─────────────────────────────┐
│ 4A. TRANSCRIPTION           │    │ 4B. AUDIO ANALYSIS          │
│                             │    │                             │
│ AssemblyAI                  │    │ Analyse du bruit            │
│                             │    │ + détection des silences    │
│ audio.opus                  │    │                             │
│      ↓                      │    │ audio.opus                  │
│ transcript                  │    │      ↓                      │
│ + timestamps par mot        │    │ noise floor                │
│   (SOURCE TIME)             │    │ threshold Auto              │
│ + confidence                │    │ DetectedSilence[] (brut)    │
│                             │    │ SourceTime                  │
│ Stocke :                    │    │                             │
│ • transcript                │    │ noise_floor_db              │
│ • transcript_word           │    │ calculated_threshold_db     │
│ • SRT source (toujours)     │    │                             │
│                             │    │ Filtres (min / paddings)    │
│ PAS d'ASS final ici         │    │ → plus tard EditDecision    │
│ (après Timeline)            │    │                             │
└──────────────┬──────────────┘    └──────────────┬──────────────┘
               │                                  │
               │ transcript.completed             │
               │ (indépendant du silence)         │
      ┌────────┴────────┐                         │
      ▼                 ▼                         │
┌───────────────┐  ┌────────────────┐             │
│ 5A. TEXT      │  │ 5B. VIRAL      │             │
│ ANALYSIS      │  │ ANALYSIS       │             │
│               │  │                │             │
│ TranscriptWord│  │ TranscriptWord │             │
│      │        │  │      ↓         │             │
│ fillers       │  │ Formatter      │             │
│ répétitions   │  │ Chunker        │             │
│ faux départs  │  │ GPT/DeepSeek   │             │
│               │  │ Boundary       │             │
│ DetectedIssue │  │ Dedup / Rank   │             │
│               │  │                │             │
│               │  │ ViralCandidate │             │
│               │  │ (SOURCE TIME)  │             │
└───────┬───────┘  └───────┬────────┘             │
        │                  │                      │
        └────────┬─────────┘                      │
                 │                                │
                 ▼                                │
┌──────────────────────────────────────────────┐
│ Analyses prêtes (silence ∥ text ∥ viral)     │
└──────────────────────────────────────────────┘
