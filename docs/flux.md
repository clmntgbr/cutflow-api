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
│ + confidence                │    │ DetectedSilence[]           │
│                             │    │                             │
│ Stocke :                    │    │                             │
│ • transcript                │    │                             │
│ • transcript_word           │    │                             │
│ • SRT source (toujours)     │    │                             │
│                             │    │                             │
│ PAS d'ASS final ici         │    │                             │
│ (après Timeline)            │    │                             │
│ TranscriptWord[]            │    │                             │
└──────────────┬──────────────┘    └──────────────┬──────────────┘
               │                                  │
               ▼                                  │
┌─────────────────────────────┐                   │
│ 5. TEXT ANALYSIS            │                   │
│                             │                   │
│ TranscriptWord[]            │                   │
│      │                      │                   │
│      ├── Fillers            │                   │
│      ├── Répétitions        │                   │
│      └── Faux départs       │                   │
│                             │                   │
│ DetectedFiller[]            │                   │
│ DetectedRepetition[]        │                   │
└──────────────┬──────────────┘                   │
               │                                  │
               ├─────────────────┐                │
               ▼                 ▼                │
┌──────────────────────────────────────────────┐
│ 6. VIRAL ANALYSIS                            │
│                                              │
│ Transcript global                           │
│       ↓                                      │
│ segmentation sémantique                      │
│       ↓                                      │
│ candidats 20-90 sec                          │
│       ↓                                      │
│ LLM                                          │
│       ↓                                      │
│ score                                        │
│ hook                                         │
│ autonomie du passage                         │
│ payoff                                       │
│       ↓                                      │
│ ViralCandidate[]                             │
└──────────────────────┬───────────────────────┘
                       │
                       │
      ┌────────────────┴────────────────┐
      │                                 │
      │ Toutes les analyses sont prêtes │
      │                                 │
      ▼                                 ▼