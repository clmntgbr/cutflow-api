# Projet --- Montage vidéo automatique (coupes, sous-titres animés, timeline, clips viraux)

Document de cadrage et d'architecture. Inspiration : AutoTrim (coupe des
silences, export timeline pour Final Cut / Premiere / Resolve) et le
style de sous-titres de la vidéo exemple (format vertical 1080x1920,
bandeau titre en haut, sous-titres en majuscules avec le mot prononcé en
surbrillance). Ce projet combine deux familles d'outils qui existent
séparément aujourd'hui : le **dérushage** (couper silences, hésitations,
répétitions) et la **création de clips courts sous-titrés**.

------------------------------------------------------------------------

## 1. Fonctionnalités visées

  -----------------------------------------------------------------------
  \#                      Fonctionnalité          Description
  ----------------------- ----------------------- -----------------------
  F1                      Upload multi-clips      Glisser-déposer
                                                  plusieurs vidéos/audios
                                                  d'un coup, traités en
                                                  parallèle

  F2                      Suppression des         Détection des blancs,
                          silences                avec réglages fins :
                                                  seuil de silence,
                                                  pré/post-roll, écart
                                                  minimum

  F3                      Suppression des mots de Détection via la
                          remplissage et          transcription (« euh »,
                          répétitions             « ben », reprises de
                                                  phrase)

  F4                      Sous-titres animés      Choix de la police,
                                                  style majuscules, mot
                                                  courant en couleur
                                                  (effet « karaoké »)

  F5                      Éléments incrustés      Bandeau/headline en
                                                  haut de la vidéo,
                                                  éventuellement logo ou
                                                  watermark

  F6                      Export timeline unique  Tous les clips traités
                                                  fusionnés en une seule
                                                  timeline FCPXML / EDL /
                                                  OTIO, prête à monter

  F7                      Export rendu            MP4 final (avec
                                                  sous-titres incrustés),
                                                  MP3, ou dossier de
                                                  clips numérotés

  F8                      Clips viraux            Détection automatique
                                                  des meilleurs moments
                                                  d'une longue vidéo et
                                                  découpe en clips
                                                  verticaux

  F9                      Préréglages             YouTube, Podcast, Vlog,
                                                  Short vertical : un
                                                  clic applique les bons
                                                  paramètres
  -----------------------------------------------------------------------

------------------------------------------------------------------------

## 2. Ce que montre la vidéo exemple

Analyse de la vidéo fournie (23 s, 1080x1920, H.264) :

-   **Format vertical 9:16**, avec deux compositions alternées : plan
    visage plein cadre, ou image d'illustration (b-roll) dans la partie
    haute avec le sous-titre en bas.
-   **Bandeau titre** en haut (rectangle blanc arrondi, texte noir),
    affiché pendant les premières secondes.
-   **Sous-titres** : gras, MAJUSCULES, contour sombre, 2 à 4 mots par
    ligne, positionnés en bas de l'écran ; **un mot est mis en couleur**
    (vert ou jaune) au moment où il est prononcé.
-   Un filigrane de l'outil d'origine est visible (OpusClip) : cette
    vidéo est un rendu d'un outil concurrent, utile comme référence de
    style, pas comme cible à copier à l'identique.

Ces trois éléments (bandeau, sous-titres mot par mot, alternance de
compositions) définissent le cahier des charges du moteur de rendu
(section 6).

------------------------------------------------------------------------

## 3. Architecture globale

``` mermaid
flowchart LR
    U[Upload multi-clips] --> ING[Ingestion]
    ING --> AUD[Extraction audio]
    AUD --> ASR[Transcription mot par mot]
    AUD --> SIL[Détection de silences]
    ASR --> FIL[Détection mots de remplissage et répétitions]
    SIL --> CUT[Moteur de décision de coupes]
    FIL --> CUT
    CUT --> TL[Timeline de coupes]
    TL --> EXP[Export timeline FCPXML / EDL / OTIO]
    TL --> REN[Rendu vidéo]
    ASR --> SUB[Génération sous-titres animés]
    SUB --> REN
    ASR --> VIR[Analyse de moments viraux - LLM]
    VIR --> CLP[Découpe en clips]
    CLP --> REN
```

Le cœur du système est **une représentation intermédiaire unique : la
timeline de coupes** (liste de segments source à conserver, avec leurs
timestamps). Tout le reste en découle : l'export vers un logiciel de
montage n'a besoin que de cette liste, le rendu vidéo l'applique en
réencodant, les clips viraux en sont simplement des sous-ensembles.

------------------------------------------------------------------------

## 4. Entités du domaine

``` mermaid
classDiagram
    class Project {
        +string ID
        +string Name
        +Preset Preset
        +ProjectStatus Status
    }

    class MediaFile {
        +string ID
        +string ProjectID
        +string StorageKey
        +Duration Duration
        +MediaFileStatus Status
    }

    class Transcript {
        +string ID
        +string MediaFileID
        +string Language
        +Word[] Words
    }

    class Word {
        +string Text
        +Duration Start
        +Duration End
        +float Confidence
        +WordKind Kind
    }

    class CutSegment {
        +string ID
        +string MediaFileID
        +Duration SourceStart
        +Duration SourceEnd
        +CutReason RemovedBefore
    }

    class Timeline {
        +string ID
        +string ProjectID
        +CutSegment[] Segments
    }

    class SubtitleStyle {
        +string FontFamily
        +int FontSize
        +string BaseColor
        +string HighlightColor
        +int WordsPerLine
        +string Position
    }

    class Overlay {
        +string ID
        +OverlayType Type
        +string Text
        +Duration Start
        +Duration End
        +string Position
    }

    class ViralClip {
        +string ID
        +string MediaFileID
        +Duration Start
        +Duration End
        +float Score
        +string Hook
    }

    Project "1" --> "*" MediaFile
    MediaFile "1" --> "0..1" Transcript
    Transcript "1" --> "*" Word
    Project "1" --> "1" Timeline
    Timeline "1" --> "*" CutSegment
    Project "1" --> "0..1" SubtitleStyle
    Project "1" --> "*" Overlay
    MediaFile "1" --> "*" ViralClip
```

  -----------------------------------------------------------------------
  Entité                              Rôle
  ----------------------------------- -----------------------------------
  `Word.Kind`                         `speech`, `filler` (« euh »),
                                      `repetition` (reprise de mot ou de
                                      phrase) : sert à décider ce qui est
                                      supprimé

  `CutSegment.RemovedBefore`          Raison de la coupe précédant ce
                                      segment (`silence`, `filler`,
                                      `repetition`), utile pour afficher
                                      pourquoi une coupe a eu lieu et
                                      permettre de l'annuler

  `SubtitleStyle`                     Police, couleur de base, couleur de
                                      surbrillance, mots par ligne,
                                      position : les réglages de la
                                      section 6

  `ViralClip.Score`                   Note de pertinence du moment
                                      (section 7), avec un `Hook` (la
                                      phrase d'accroche détectée)
  -----------------------------------------------------------------------

------------------------------------------------------------------------

## 5. Traitement audio : coupes, silences, mots de remplissage

### 5.1 Détection de silences

Traitement purement audio, très léger : `ffmpeg silencedetect` (ou
détection d'énergie/VAD) suffit. Mesure faite sur une vidéo de 5:50 :
**environ 2 secondes** pour analyser toute la piste audio. Aucun besoin
de GPU ni d'IA à cette étape.

  ---------------------------------------------------------------------------
  Paramètre                   Rôle                    Défaut proposé
  --------------------------- ----------------------- -----------------------
  `silence_threshold_db`      Niveau sous lequel on   -35 dB
                              considère qu'il y a     
                              silence                 

  `min_silence_duration_ms`   Écart minimum pour      400 ms
                              qu'un blanc soit coupé  

  `pre_roll_ms`               Marge conservée avant   100 ms
                              la reprise de parole    

  `post_roll_ms`              Marge conservée après   150 ms
                              la fin de parole        
  ---------------------------------------------------------------------------

Le pré/post-roll évite les coupes « sèches » qui rognent le début ou la
fin des mots : c'est le réglage qui fait la différence entre un montage
naturel et un montage haché.

### 5.2 Transcription mot par mot

Nécessaire pour les mots de remplissage, les répétitions, **et** les
sous-titres animés. Il faut des timestamps **au niveau du mot**
(début/fin de chaque mot), pas seulement par phrase. C'est l'étape la
plus coûteuse en calcul du projet (section 8).

### 5.3 Mots de remplissage et répétitions

-   **Remplissage** : dictionnaire par langue (« euh », « hum », « ben
    », « genre », « en fait » selon le contexte), à appliquer sur la
    transcription. Attention : ces mots sont parfois légitimes (« en
    fait » qui porte du sens) → à proposer comme option réglable par
    niveau d'agressivité, et à marquer comme expérimental au départ,
    comme le fait AutoTrim.
-   **Répétitions** : détection de séquences de mots identiques ou quasi
    identiques consécutives (« je je pense », « on va, on va voir ») ;
    garder la dernière occurrence, en général la plus fluide.
-   **Faux départs** (phrase reprise entièrement) : plus difficile, à
    traiter en second temps.

### 5.4 Moteur de décision de coupes

Fusionne les trois sources (silences, remplissage, répétitions) en une
seule liste de segments à conserver :

    segments_a_garder = durée_totale
                        − silences (avec pré/post-roll)
                        − mots_de_remplissage_marqués
                        − répétitions_marquées
    puis : fusionner les segments séparés par moins de X ms (éviter les micro-coupes)

Chaque coupe garde sa raison (`RemovedBefore`) pour l'affichage dans
l'aperçu et pour pouvoir la réactiver individuellement.

------------------------------------------------------------------------

## 6. Sous-titres animés et incrustations

### 6.1 Génération des sous-titres mot par mot

À partir de la transcription avec timestamps par mot :

1.  Regrouper les mots en lignes de 2 à 4 mots (paramètre
    `words_per_line`), en coupant aux pauses naturelles.
2.  Pour chaque ligne, produire une suite d'états : le mot courant en
    `highlight_color`, les autres en `base_color`.
3.  Formats de sortie possibles :
    -   **ASS/SSA** (sous-titres avancés) : gère nativement le style par
        mot, le contour, la police, la position ; rendu direct par
        ffmpeg (`libass`). C'est la voie la plus simple pour un premier
        rendu.
    -   **Rendu par calques** (moteur de composition type Remotion ou
        Skia) : plus de liberté graphique (animations, rebond du mot,
        transitions), plus lourd à mettre en place.

  -----------------------------------------------------------------------
  Réglage utilisateur                 Effet
  ----------------------------------- -----------------------------------
  Police                              Liste de polices embarquées
                                      (licences à vérifier pour un usage
                                      commercial)

  Couleur de base / de surbrillance   Comme la vidéo exemple : blanc +
                                      vert/jaune

  Mots par ligne                      2 à 4

  Position                            Bas, milieu, haut

  Contour / ombre                     Lisibilité sur fond variable
  -----------------------------------------------------------------------

### 6.2 Incrustations (bandeau titre)

Éléments simples posés sur la timeline : un texte dans un rectangle
arrondi (comme le bandeau blanc en haut de la vidéo exemple), avec
début/fin, position, style. Modèle `Overlay` (section 4). Rendu par le
même moteur que les sous-titres pour garder un seul pipeline graphique.

### 6.3 Recadrage vertical

Pour les clips viraux au format 9:16 à partir d'une vidéo horizontale :
recadrage centré sur le visage (détection de visage, suivi du locuteur),
ou composition « split » (visage en bas, illustration en haut) comme
dans la vidéo exemple. Fonctionnalité à part entière, à traiter après le
MVP.

------------------------------------------------------------------------

## 7. Détection de moments viraux

Pipeline en trois temps, à partir de la transcription (pas de la vidéo)
:

``` mermaid
flowchart LR
    A[Transcription complète] --> B[Découpage en segments candidats de 20 à 90 s]
    B --> C[Notation par LLM: accroche, émotion, autonomie du propos, chute]
    C --> D[Classement et sélection des N meilleurs]
    D --> E[Ajustement des bornes sur les phrases entières]
    E --> F[Clip vertical sous-titré]
```

  -----------------------------------------------------------------------
  Critère de notation                 Ce que cherche le LLM
  ----------------------------------- -----------------------------------
  Accroche                            Les premières secondes donnent une
                                      raison de rester (question,
                                      affirmation forte)

  Autonomie                           Le passage se comprend sans le
                                      reste de la vidéo

  Émotion / intensité                 Surprise, humour, controverse,
                                      révélation

  Chute                               Le passage se termine sur une
                                      conclusion, pas au milieu d'une
                                      idée
  -----------------------------------------------------------------------

Points clés : - **Travailler sur du texte est peu coûteux** : un LLM sur
une transcription d'1 heure coûte très peu comparé au traitement
vidéo. - **Les bornes du clip doivent tomber sur des fins de phrase**
(le timestamp mot par mot permet de le faire proprement), sinon le clip
commence ou finit au milieu d'un mot. - **Le score est une aide, pas une
vérité** : présenter les clips candidats à l'utilisateur avec la raison
du score, pour qu'il valide ou ajuste. - Il n'existe pas de mesure
fiable de « viralité » : le système estime un potentiel à partir de
critères rédactionnels, il ne peut pas prédire les vues.

------------------------------------------------------------------------

## 8. Export : timeline vs rendu

C'est la distinction économique la plus importante du projet :

  -----------------------------------------------------------------------
  Sortie                  Coût de calcul          Explication
  ----------------------- ----------------------- -----------------------
  **Export timeline**     Quasi nul               On ne modifie pas la
  (FCPXML, EDL, OTIO)                             vidéo, on écrit un
                                                  fichier texte listant
                                                  les coupes (références
                                                  vers les fichiers
                                                  source, timestamps
                                                  début/fin)

  **Rendu MP4** avec      Modéré                  Réencodage vidéo (ou
  coupes                                          copie de flux si les
                                                  coupes tombent sur des
                                                  keyframes, avec perte
                                                  de précision)

  **Rendu MP4** avec      Élevé                   Réencodage complet de
  sous-titres incrustés                           la vidéo avec
                                                  incrustation des
                                                  sous-titres et du
                                                  bandeau

  **Clips viraux**        Élevé                   Recadrage +
  verticaux sous-titrés                           sous-titres +
                                                  réencodage, pour chaque
                                                  clip
  -----------------------------------------------------------------------

Formats de timeline : - **FCPXML** : Final Cut Pro (et importable dans
DaVinci Resolve). - **EDL (CMX 3600)** : format ancien et universel,
limité (pas de métadonnées riches), lu par presque tous les logiciels. -
**OpenTimelineIO (OTIO)** : format ouvert, avec des adaptateurs vers
plusieurs outils ; pratique comme représentation interne d'où dériver
les autres exports. - **XML Premiere** (`xmeml`) : à prévoir séparément,
format différent de FCPXML. - **CapCut** n'importe pas de timeline :
export en **dossier de clips numérotés** (001, 002, 003...) dans
l'ordre, comme le fait AutoTrim.

**Les sous-titres animés ne passent généralement pas dans une timeline
XML** : les styles par mot ne sont pas portés de façon fiable d'un
logiciel à l'autre. Deux options : exporter un fichier de sous-titres
standard (SRT/ASS) à côté de la timeline, ou proposer le rendu MP4 avec
sous-titres incrustés en complément.

------------------------------------------------------------------------

## 9. Modèle de déploiement : local ou cloud

C'est la décision structurante, à trancher avant de coder :

  -----------------------------------------------------------------------
                          Application locale      SaaS cloud
                          (type AutoTrim)         
  ----------------------- ----------------------- -----------------------
  Où tourne le calcul     Sur la machine de       Sur tes serveurs
                          l'utilisateur (CPU/GPU) 

  Coût par utilisateur    Quasi nul               Proportionnel à l'usage
  pour toi                                        (transcription, rendu)

  Modèle économique       Paiement unique ou      Abonnement avec quotas
  naturel                 licence                 

  Confidentialité         Fort : les fichiers ne  À justifier (uploads de
  (argument de vente)     quittent pas la machine vidéos brutes)

  Vitesse perçue          Pas d'upload, mais      Upload de gros
                          dépend de la machine    fichiers, puis
                                                  traitement rapide sur
                                                  serveurs

  Complexité de           Installeurs             Web, mises à jour
  distribution            macOS/Windows, mises à  instantanées
                          jour, licences          

  Difficulté technique    Modèles IA à embarquer  Infra GPU à gérer et à
                          et optimiser par        dimensionner
                          plateforme              
  -----------------------------------------------------------------------

AutoTrim se positionne explicitement sur le local (« 100 % local et
privé », paiement unique, aucun coût serveur). Un SaaS cloud sur les
mêmes fonctionnalités de dérushage entre en concurrence sur un terrain
où l'adversaire a un coût marginal nul.

**Où le cloud garde un vrai avantage** : les fonctionnalités qui
demandent une puissance que l'utilisateur n'a pas ou qu'il ne veut pas
gérer (transcription de qualité sur des heures de contenu, analyse
virale par LLM, rendu de nombreux clips verticaux), et l'usage depuis un
navigateur ou un téléphone.

**Option hybride à considérer** : application locale pour le dérushage
(silences, timeline), services cloud facturés à l'usage pour la
transcription et les clips viraux.

------------------------------------------------------------------------

## 10. Coûts et scalabilité (à benchmarker avant de s'engager)

Aucun chiffre ci-dessous n'est mesuré sur ce projet, sauf la détection
de silences ; ce sont les postes de coût à quantifier en priorité :

  -----------------------------------------------------------------------
  Poste                   Nature du coût          Ordre de grandeur
                                                  relatif
  ----------------------- ----------------------- -----------------------
  Détection de silences   CPU, audio seul         Négligeable (\~2 s pour
                                                  5:50 mesuré)

  Export timeline         Écriture de fichier     Négligeable

  Transcription mot par   Modèle de               **Poste principal** du
  mot                     reconnaissance vocale   traitement audio : à
                          (GPU recommandé, ou     mesurer en temps de
                          API)                    traitement par minute
                                                  d'audio

  Analyse virale          Appel LLM sur du texte  Faible à modéré

  Rendu vidéo             Réencodage CPU/GPU      **Poste principal**
  (sous-titres incrustés,                         côté vidéo, croît avec
  clips verticaux)                                le nombre de minutes
                                                  rendues

  Stockage et bande       Vidéos brutes et rendus Croît vite avec les
  passante                                        fichiers volumineux ;
                                                  prévoir une politique
                                                  de
                                                  rétention/suppression
  -----------------------------------------------------------------------

Leçon tirée du projet précédent (scan de vidéos par OCR image par image)
: **le coût se joue sur le volume de données traitées par utilisateur**,
et il faut le mesurer sur des vidéos représentatives avant de
dimensionner l'infrastructure ou de fixer un prix. Ici, l'audio est bien
plus léger que l'image (une piste audio pèse peu, la transcription est
un seul passage), ce qui joue en faveur du projet : la partie «
dérushage + timeline » ne dépend pas du traitement image par image.

------------------------------------------------------------------------

## 11. Concurrence et positionnement

  -----------------------------------------------------------------------
  Outil                               Positionnement connu
  ----------------------------------- -----------------------------------
  AutoTrim                            Dérushage local, suppression de
                                      silences/hésitations, export
                                      timeline FCP/Premiere/Resolve,
                                      paiement unique

  OpusClip                            Clips viraux verticaux sous-titrés
                                      à partir de longues vidéos (source
                                      du style de la vidéo exemple)

  Descript                            Montage par édition de la
                                      transcription

  TimeBolt, AutoCut                   Suppression de silences pour
                                      monteurs
  -----------------------------------------------------------------------

Le marché est occupé sur chaque brique prise séparément. Ce qui peut
différencier : **le pipeline complet dans un seul outil** (dérusher →
sous-titrer → sortir en timeline **et** en clips verticaux), ou une
**spécialisation** (une langue, un type de contenu comme le podcast ou
la formation, un public comme les agences). À valider avec de vrais
utilisateurs avant de construire l'ensemble : la question à poser est
laquelle des trois briques ils sont prêts à payer, pas s'ils les
trouvent intéressantes.

------------------------------------------------------------------------

## 12. Feuille de route proposée

  --------------------------------------------------------------------------
  Phase                   Contenu                    Pourquoi dans cet ordre
  ----------------------- -------------------------- -----------------------
  **MVP**                 Upload, détection de       Zéro IA, coût quasi
                          silences, réglages (seuil, nul, valeur immédiate
                          pré/post-roll, écart       pour les monteurs :
                          minimum), export           permet de tester la
                          FCPXML/EDL, aperçu des     demande avant
                          coupes                     d'investir dans le
                                                     reste

  **V1**                  Transcription mot par mot, Ajoute la brique IA
                          suppression des mots de    principale, débloque
                          remplissage/répétitions,   les sous-titres
                          export SRT/ASS             

  **V2**                  Sous-titres animés         Ajoute le rendu, le
                          incrustés (rendu MP4),     poste le plus coûteux :
                          bandeau titre, polices et  à faire une fois la
                          couleurs configurables     demande validée

  **V3**                  Clips viraux, recadrage    Fonctionnalité la plus
                          vertical automatique       concurrentielle
                                                     (OpusClip), la plus
                                                     lourde ; à faire en
                                                     dernier
  --------------------------------------------------------------------------

------------------------------------------------------------------------

## 13. Points d'attention

-   **Précision des coupes** : une coupe qui rogne un mot est
    immédiatement visible. Le pré/post-roll et un aperçu écoutable de
    chaque coupe sont indispensables, pas optionnels.
-   **Mots de remplissage multilingues** : les listes et la qualité de
    détection varient beaucoup selon la langue ; démarrer sur une seule
    langue (le français) et l'étiqueter comme expérimental.
-   **Droits des polices** : les polices proposées pour les sous-titres
    doivent avoir une licence compatible avec un usage commercial et
    l'incrustation dans des vidéos.
-   **Droits sur les contenus** : les clips viraux sont tirés de vidéos
    que l'utilisateur doit avoir le droit de réutiliser ; prévoir une
    clause dans les conditions d'utilisation.
-   **Fichiers volumineux** : les vidéos brutes font souvent plusieurs
    gigaoctets ; upload reprenable (multipart), URL présignées et
    politique de suppression automatique à définir dès le début
    (cf. l'architecture d'upload déjà conçue pour le projet précédent,
    réutilisable telle quelle).
-   **Réutilisation de l'existant** : la base CQRS/event-driven Go,
    RabbitMQ, le stockage S3/MinIO, les statuts par étape et le
    découpage en workers parallèles sont directement réutilisables ;
    seuls les workers changent (transcription, décision de coupes, rendu
    à la place d'OCR).

------------------------------------------------------------------------

## 14. Complément d'architecture --- pipeline parallèle, Master Render et clips viraux

### 14.1 Pas de TechnicalSegment pour silence / ASR

Avec Opus (fichier audio léger) et AssemblyAI (ASR full-file + SRT),
le découpage en `TechnicalSegment` n'apporte rien pour la détection
de silences ni la transcription :

- `ffmpeg silencedetect` sur un Opus d'1 h reste négligeable ;
- AssemblyAI accepte le fichier complet et renvoie SRT + mots horodatés ;
- chunking + overlap + merge ajouterait de la complexité sans gain
  de latence ni de coût measurable à notre échelle.

On traite donc **l'audio complet** en parallèle après extraction :

``` text
VIDEO SOURCE
    ↓
EXTRACTION AUDIO (Opus)
    ↓
media_file.audio_ready
    ├── media_file.silence_requested  → worker silence (ffmpeg)
    └── media_file.transcript_requested → worker transcript (AssemblyAI → SRT → ASS)
```

Les `CutSegment` (montage) et `ViralClip` restent distincts et viendront
plus tard à partir du transcript global + des silences détectés.

### 14.2 Extraction audio unique

L'audio est extrait **une seule fois** par média source (worker
`extraction` sur `media_file.ready.v1`). Silence et transcription
réutilisent la même clé MinIO.

### 14.3 Transcript global (full-file)

Pas de `TranscriptSegment` / merge d'overlap : un seul job AssemblyAI
par média produit :

- `transcript.text` + `transcript_word[]` (timestamps source) ;
- `subtitles.srt` (API AssemblyAI, fallback plat).

L'ASS animé (highlight mot actif) n'est **pas** produit par le provider.
Il est généré localement par le `SubtitleGenerator` à partir de
`TranscriptWord[]` + `SubtitleStyle` (preset TikTok Classic par défaut
aujourd'hui). Le remapping `SourceTime → OutputTime` s'appliquera plus
tard via la Timeline avant régénération ASS au rendu.

Le transcript global sert ensuite à détecter fillers / répétitions,
régénérer les sous-titres (plusieurs presets), et analyser les moments
viraux.

### 14.4 Analyse virale globale

La détection des moments viraux s'appuie sur le **transcript global**,
pas sur des fenêtres techniques. Un passage intéressant est découpé
sémantiquement (20–90 s) puis scoré par LLM.

``` text
GLOBAL TRANSCRIPT
       ↓
segmentation sémantique
       ↓
candidats de 20–90 s
       ↓
LLM
       ↓
ViralClip[]
```

Les bornes des clips viraux sont donc connues avant le rendu final.

### 14.5 Timeline comme source de vérité

La `Timeline` est la représentation centrale du montage.

Elle détermine :

-   les portions conservées ;
-   les portions supprimées ;
-   l'ordre des médias ;
-   le passage de `SourceTime` vers `OutputTime` ;
-   les exports FCPXML / EDL / OTIO ;
-   les timestamps des sous-titres ;
-   le rendu final.

Il faut distinguer :

``` text
SOURCE TIME
00:00 ------------------------------ 30:00
              ↓ coupes
OUTPUT TIME
00:00 ------------------------ 24:32
```

### 14.6 ASS/SRT après remapping

Les fichiers de sous-titres finaux doivent être générés à partir du
transcript global **et de la timeline finale**.

``` text
GLOBAL TRANSCRIPT
       +
FINAL TIMELINE
       ↓
TIMESTAMP REMAPPING
       ├── SRT
       └── ASS
```

Une coupe de cinq secondes décale tous les sous-titres situés après
cette coupe dans la vidéo de sortie.

### 14.7 Conversion 16:9 → 9:16

Le passage au format TikTok / Reels / Shorts ne nécessite pas
obligatoirement de tracking de visage.

Les modes simples peuvent être :

-   crop centré ;
-   crop gauche ;
-   crop droite ;
-   crop personnalisé ;
-   vidéo entière avec fond flouté.

Le crop et le scale nécessitent un réencodage, mais pas d'analyse IA
image par image.

### 14.8 Un seul rendu final

Éviter :

``` text
original.mp4
  ↓
cut.mp4
  ↓
vertical.mp4
  ↓
subtitles.mp4
  ↓
overlay.mp4
  ↓
final.mp4
```

Construire plutôt un seul pipeline FFmpeg :

``` text
SOURCE
  ↓
trim / atrim
  ↓
concat
  ↓
crop
  ↓
scale
  ↓
ASS
  ↓
overlays
  ↓
ENCODE
  ↓
FINAL.MP4
```

Le but est de ne réencoder la vidéo qu'une seule fois.

### 14.9 Master Render

Un `MasterRender` est une vidéo complètement rendue correspondant à une
timeline et à un ensemble précis de paramètres :

-   format ;
-   résolution ;
-   crop ;
-   style de sous-titres ;
-   overlays ;
-   codec ;
-   framerate.

Exemple :

``` text
Timeline finale
      +
Preset Social 9:16
      +
ASS
      +
Overlays
      ↓
MASTER-SOCIAL.MP4
```

### 14.10 Extraction rapide des clips viraux

Si un clip viral utilise exactement le même format et le même style que
le Master Render, il n'est pas nécessaire de le réencoder.

Exemple :

``` bash
ffmpeg -ss 134.2 -i master.mp4 -t 38.6 -c copy viral-01.mp4
```

Avec `-c copy`, FFmpeg copie les flux sans décoder/réencoder toute la
vidéo. Le coût CPU devient très faible.

### 14.11 Keyframes

Le stream copy dépend des keyframes pour obtenir des coupes propres.

Puisque les `ViralClip` sont connus avant le Master Render, le renderer
peut forcer des keyframes aux débuts des clips sélectionnés :

``` text
ViralClip 1 start → 134.200 s
ViralClip 2 start → 512.500 s
ViralClip 3 start → 1024.800 s
```

Ces timestamps peuvent être transmis au renderer comme keyframes
forcées.

### 14.12 Réutilisation du Master

Un clip peut utiliser `stream_copy` lorsque ses paramètres correspondent
au Master :

``` text
Master Social
1080×1920
9:16
H.264
crop=center
subtitle_style=A
logo=A
```

et :

``` text
ViralClip
1080×1920
9:16
H.264
crop=center
subtitle_style=A
logo=A
```

Dans ce cas :

``` text
MASTER
  ↓
STREAM COPY
  ↓
VIRAL CLIP
```

Si le Master est 16:9 et le clip demandé en 9:16, ou si le
style/crop/overlay diffère, un nouveau rendu est nécessaire.

### 14.13 Plusieurs Masters

Un projet peut avoir plusieurs Masters lorsque cela est réellement utile
:

``` text
PROJECT
   ├── Master YouTube 16:9
   └── Master Social 9:16
          ├── TikTok-01.mp4
          ├── TikTok-02.mp4
          ├── Reel-01.mp4
          └── Short-01.mp4
```

Les clips sociaux peuvent alors être extraits du Master Social sans
rendu complet supplémentaire.

### 14.14 Pipeline final

``` text
                    VIDEO SOURCE
                         │
                         ▼
                  INGEST / FFPROBE
                         │
                ┌────────┴────────┐
                ▼                 ▼
         EXTRACTION AUDIO   TECHNICAL SEGMENTS
                │
                ▼
        AUDIO SEGMENTS
                │
       ┌────────┴─────────┐
       ▼                  ▼
SILENCE DETECTION   TRANSCRIPTION
       │                  │
       │                  ▼
       │          TRANSCRIPT MERGE
       │                  │
       │        ┌─────────┼──────────┐
       │        ▼         ▼          ▼
       │     FILLERS  REPETITIONS  VIRAL
       │        │         │          │
       └────────┴─────────┴──────────┘
                         │
                         ▼
                  TIMELINE ENGINE
                         │
                         ▼
                  FINAL TIMELINE
                         │
              ┌──────────┼─────────┐
              ▼          ▼         ▼
           FCPXML       ASS       SRT
                         │
                         ▼
                 USER VALIDATION
                         │
                         ▼
                    RENDER PLAN
                         │
               ┌─────────┴─────────┐
               ▼                   ▼
        MASTER YOUTUBE       MASTER SOCIAL
            16:9                  9:16
               │                   │
               │             stream copy
               │            ┌──────┼──────┐
               │            ▼      ▼      ▼
               │          Clip1  Clip2  Clip3
               ▼
          YouTube.mp4
```

------------------------------------------------------------------------

## 15. DDL PostgreSQL

Les timestamps média sont stockés en millisecondes (`BIGINT`). Cela
simplifie les calculs de timeline, les échanges avec FFmpeg et le
remapping des sous-titres.

``` sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE project_status AS ENUM (
    'draft', 'processing', 'ready', 'rendering', 'completed', 'failed'
);

CREATE TYPE media_status AS ENUM (
    'uploaded', 'probing', 'ready', 'processing', 'completed', 'failed'
);

CREATE TYPE job_status AS ENUM (
    'pending', 'queued', 'processing', 'completed', 'failed', 'cancelled'
);

CREATE TYPE word_kind AS ENUM ('speech', 'filler', 'repetition');

CREATE TYPE cut_reason AS ENUM (
    'silence', 'filler', 'repetition', 'manual'
);

CREATE TYPE render_type AS ENUM (
    'master', 'viral_clip', 'custom'
);

CREATE TYPE render_strategy AS ENUM (
    'full_render', 'stream_copy'
);

CREATE TYPE aspect_ratio AS ENUM (
    '16:9', '9:16', '1:1', '4:5', 'custom'
);

CREATE TYPE crop_mode AS ENUM (
    'none', 'center', 'left', 'right', 'custom', 'fit_blur'
);

CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    status project_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE media_file (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL,
    original_filename TEXT,
    mime_type VARCHAR(100),
    duration_ms BIGINT NOT NULL,
    width INTEGER,
    height INTEGER,
    fps NUMERIC(8,3),
    video_codec VARCHAR(50),
    audio_codec VARCHAR(50),
    size_bytes BIGINT,
    status media_status NOT NULL DEFAULT 'uploaded',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (duration_ms >= 0)
);

CREATE INDEX idx_media_file_project ON media_file(project_id);

CREATE TABLE media_audio (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL UNIQUE REFERENCES media_file(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL,
    codec VARCHAR(50),
    sample_rate INTEGER,
    channels SMALLINT,
    size_bytes BIGINT,
    status job_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pas de technical_segment : silence + ASR full-file sur l'audio extrait.

CREATE TABLE detected_silence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    start_ms BIGINT NOT NULL,
    end_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_ms >= 0),
    CHECK (end_ms > start_ms)
);

CREATE INDEX idx_detected_silence_media_time
    ON detected_silence(media_file_id, start_ms);

CREATE TABLE transcript (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL UNIQUE REFERENCES media_file(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10),
    text TEXT,
    srt_storage_key TEXT NOT NULL DEFAULT '',
    ass_storage_key TEXT NOT NULL DEFAULT '',
    provider_job_id VARCHAR(100),
    status job_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transcript_word (
    id BIGSERIAL PRIMARY KEY,
    transcript_id UUID NOT NULL REFERENCES transcript(id) ON DELETE CASCADE,
    word_index INTEGER NOT NULL,
    text TEXT NOT NULL,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    confidence NUMERIC(5,4),
    kind word_kind NOT NULL DEFAULT 'speech',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms >= source_start_ms),
    CHECK (confidence IS NULL OR confidence BETWEEN 0 AND 1),
    UNIQUE (transcript_id, word_index)
);

CREATE INDEX idx_transcript_word_time
    ON transcript_word(transcript_id, source_start_ms);

CREATE INDEX idx_transcript_word_kind
    ON transcript_word(transcript_id, kind);

CREATE TABLE timeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    duration_ms BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, version)
);

CREATE INDEX idx_timeline_project_active
    ON timeline(project_id, is_active);

CREATE TABLE cut_segment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timeline_id UUID NOT NULL REFERENCES timeline(id) ON DELETE CASCADE,
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    segment_index INTEGER NOT NULL,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    output_start_ms BIGINT NOT NULL,
    output_end_ms BIGINT NOT NULL,
    removed_before_reason cut_reason,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms > source_start_ms),
    CHECK (output_start_ms >= 0),
    CHECK (output_end_ms > output_start_ms),
    UNIQUE (timeline_id, segment_index)
);

CREATE INDEX idx_cut_segment_timeline
    ON cut_segment(timeline_id, segment_index);

CREATE INDEX idx_cut_segment_source
    ON cut_segment(media_file_id, source_start_ms);

CREATE TABLE viral_clip (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    timeline_id UUID NOT NULL REFERENCES timeline(id) ON DELETE CASCADE,
    media_file_id UUID REFERENCES media_file(id) ON DELETE SET NULL,
    source_start_ms BIGINT,
    source_end_ms BIGINT,
    output_start_ms BIGINT NOT NULL,
    output_end_ms BIGINT NOT NULL,
    score NUMERIC(5,4),
    hook TEXT,
    reason TEXT,
    selected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (output_start_ms >= 0),
    CHECK (output_end_ms > output_start_ms),
    CHECK (score IS NULL OR score BETWEEN 0 AND 1)
);

CREATE INDEX idx_viral_clip_project ON viral_clip(project_id);
CREATE INDEX idx_viral_clip_timeline_time
    ON viral_clip(timeline_id, output_start_ms);

CREATE INDEX idx_viral_clip_selected
    ON viral_clip(project_id, selected)
    WHERE selected = TRUE;

CREATE TABLE subtitle_style (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    font_family VARCHAR(255),
    font_size INTEGER,
    base_color VARCHAR(20),
    highlight_color VARCHAR(20),
    outline_color VARCHAR(20),
    outline_size INTEGER,
    words_per_line SMALLINT,
    position VARCHAR(20),
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE overlay (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    text TEXT,
    storage_key TEXT,
    start_ms BIGINT NOT NULL,
    end_ms BIGINT NOT NULL,
    position VARCHAR(50),
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_ms >= 0),
    CHECK (end_ms > start_ms)
);

CREATE TABLE output_preset (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    aspect_ratio aspect_ratio NOT NULL,
    crop_mode crop_mode NOT NULL DEFAULT 'none',
    video_codec VARCHAR(50) NOT NULL DEFAULT 'h264',
    audio_codec VARCHAR(50) NOT NULL DEFAULT 'aac',
    fps NUMERIC(8,3),
    video_bitrate BIGINT,
    audio_bitrate BIGINT,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE master_render (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    timeline_id UUID NOT NULL REFERENCES timeline(id) ON DELETE RESTRICT,
    output_preset_id UUID NOT NULL REFERENCES output_preset(id) ON DELETE RESTRICT,
    subtitle_style_id UUID REFERENCES subtitle_style(id) ON DELETE SET NULL,
    storage_key TEXT,
    duration_ms BIGINT,
    status job_status NOT NULL DEFAULT 'pending',
    size_bytes BIGINT,
    render_fingerprint VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE (timeline_id, output_preset_id, render_fingerprint)
);

CREATE INDEX idx_master_render_project ON master_render(project_id);
CREATE INDEX idx_master_render_status ON master_render(status);

CREATE TABLE master_render_keyframe (
    master_render_id UUID NOT NULL REFERENCES master_render(id) ON DELETE CASCADE,
    timestamp_ms BIGINT NOT NULL,
    viral_clip_id UUID REFERENCES viral_clip(id) ON DELETE SET NULL,
    PRIMARY KEY (master_render_id, timestamp_ms),
    CHECK (timestamp_ms >= 0)
);

CREATE TABLE render_job (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    timeline_id UUID NOT NULL REFERENCES timeline(id) ON DELETE RESTRICT,
    type render_type NOT NULL,
    strategy render_strategy NOT NULL,
    output_preset_id UUID NOT NULL REFERENCES output_preset(id) ON DELETE RESTRICT,
    master_render_id UUID REFERENCES master_render(id) ON DELETE SET NULL,
    viral_clip_id UUID REFERENCES viral_clip(id) ON DELETE SET NULL,
    start_ms BIGINT,
    end_ms BIGINT,
    storage_key TEXT,
    status job_status NOT NULL DEFAULT 'pending',
    progress NUMERIC(5,2) NOT NULL DEFAULT 0,
    error_message TEXT,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CHECK (progress >= 0 AND progress <= 100),
    CHECK (start_ms IS NULL OR start_ms >= 0),
    CHECK (end_ms IS NULL OR start_ms IS NULL OR end_ms > start_ms)
);

CREATE INDEX idx_render_job_status ON render_job(status);
CREATE INDEX idx_render_job_project ON render_job(project_id);
CREATE INDEX idx_render_job_master ON render_job(master_render_id);
CREATE INDEX idx_render_job_viral_clip ON render_job(viral_clip_id);

CREATE TABLE processing_job (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    media_file_id UUID REFERENCES media_file(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    status job_status NOT NULL DEFAULT 'pending',
    attempt INTEGER NOT NULL DEFAULT 0,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT,
    queued_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_processing_job_status_type
    ON processing_job(status, type);

CREATE INDEX idx_processing_job_project
    ON processing_job(project_id);
```

------------------------------------------------------------------------

## 16. Principes définitifs

1.  Les fichiers sources sont immuables.
2.  Silence et ASR travaillent sur l'audio complet (pas de TechnicalSegment).
3.  L'audio est extrait une seule fois.
4.  Silence et transcription sont parallélisés (fan-out outbox).
5.  Un seul transcript global par média (AssemblyAI → SRT → ASS).
6.  L'analyse virale travaille sur le transcript global.
7.  Les moments viraux sont déterminés avant le rendu.
8.  La Timeline est la source de vérité.
9.  `SourceTime` et `OutputTime` sont distincts.
10. ASS/SRT de montage sont générés après remapping sur la timeline finale.
11. L'utilisateur valide avant le rendu coûteux.
12. Crop, scale, ASS et overlays sont appliqués dans un seul pipeline
    FFmpeg.
13. Un Master ne subit idéalement qu'un seul réencodage.
14. Les keyframes des clips viraux peuvent être forcées pendant le
    Master Render.
15. Les clips compatibles avec le Master sont extraits en `stream copy`.
16. Un nouveau rendu n'est effectué que lorsque le clip demande un
    format ou un style différent.
17. Plusieurs sorties indépendantes peuvent être traitées en parallèle.
18. Le coût principal du système est exprimé en **minutes transcrites +
    minutes rendues**.
