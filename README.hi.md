# Gommit [![Go Report Card](https://goreportcard.com/badge/github.com/Hangell/gommit)](https://goreportcard.com/report/github.com/Hangell/gommit) [![GitHub tag](https://img.shields.io/github/v/tag/Hangell/gommit?label=version&color=orange)](https://github.com/Hangell/gommit/tags) [![Build](https://github.com/Hangell/gommit/actions/workflows/ci.yml/badge.svg)](https://github.com/Hangell/gommit/actions/workflows/ci.yml) [![License](https://img.shields.io/github/license/Hangell/gommit)](LICENSE) [![Contributing](https://img.shields.io/badge/contributions-welcome-brightgreen)](CONTRIBUTING.md) [![Go Reference](https://pkg.go.dev/badge/github.com/Hangell/gommit.svg)](https://pkg.go.dev/github.com/Hangell/gommit)

**gommit** Go में लिखा गया एक तेज़, **शून्य‑निर्भरता** CLI सहायक है जो _Conventional Commits_ संदेश तैयार करता है।  
यह Commitizen/cz जैसा इंटरैक्टिव विज़ार्ड खोलता है और सही फ़ॉर्मेट के साथ `git commit` चलाता है।

> 💡 डिफ़ॉल्ट रूप से **टाइप इमोजी** **हेडर** में जोड़ा जाता है (जैसे `feat 💡: ...`) ताकि वह GitHub की फ़ोल्डर/फ़ाइल सूचियों में दिखे। इमोजी केवल शीर्षक में जोड़े जाते हैं, बॉडी/फ़ूटर में नहीं।

---

## ✨ फीचर्स

- ✅ इंटरैक्टिव विज़ार्ड, सर्च/शॉर्टकट (नंबर, कीवर्ड, `q` से बाहर)
- ✅ स्टैंडर्ड Conventional Commits प्रकार: **feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert**  
  अतिरिक्त: **WIP, prune**
- ✅ मेन्यू इमोजी (ऑटो ASCII फॉलबैक) और **हेडर इमोजी**
- ✅ Git प्री‑चेक्स:
  - रेपो वैलिडेशन (`not a git repository…` अगर Git repo नहीं है)
  - जब कुछ भी staged नहीं हो तो Auto‑stage (`git add -A`) (टॉगल)
  - कमिट से पहले staged changes सूची (टॉगल)
  - `--amend` मोड, अंतिम कमिट का बैनर दिखाता है
- ✅ Commitizen‑स्टाइल फ़ील्ड्स:
  - `scope` (वैकल्पिक)
  - `subject` (आवश्यक, `<=72` अक्षर)
  - मल्टीलाइन `body` (समाप्ति **एकल `.` लाइन** से **या** **दो बार Enter**)
  - `BREAKING CHANGE` (प्रॉम्प्ट और विवरण)
  - Issues: `Closes #123` / `Refs #45` (यदि चाहें)
- ✅ बिल्ट‑इन **`--install`** मोड — यूज़र PATH में बायनरी इंस्टॉल/अपडेट
  - Windows: `%LOCALAPPDATA%\Programs\gommitin`
  - Linux/macOS: `~/.local/bin` (ज़रूरत हो तो PATH में जोड़ता है)

---

## 🚀 इंस्टॉलेशन

### तरीका 1: एक‑लाइन स्क्रिप्ट **(सुझाया गया)**

**Windows:**
```powershell
powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"
```

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
```

### तरीका 2: Release से बायनरी डाउनलोड
1. **Releases** पेज से अपने OS/आर्क का पैकेज डाउनलोड करें
2. पैकेज निकालें (`.zip` Windows / `.tar.gz` Linux/macOS)
3. निकले हुए बायनरी को `--install` के साथ चलाएँ:

**Windows (PowerShell):**
```powershell
.\gommit.exe --install
# टर्मिनल बंद कर फिर खोलें
gommit --version
```

**Linux/macOS:**
```bash
chmod +x ./gommit
./gommit --install
# शेल फिर खोलें
gommit --version
```

> `--install` दोबारा चलाने पर इंस्टॉल की गई वर्ज़न अपडेट हो जाएगी।

---

## 💻 त्वरित उपयोग

### इंटरफ़ेस भाषा

कमांड और flags अंग्रेज़ी में ही रहते हैं। मेनू, विवरण, प्रश्न और `--help` अंग्रेज़ी,
स्पेनिश, पुर्तगाली, हिंदी, रूसी या चीनी में दिखाए जा सकते हैं:

```bash
gommit --language hi          # केवल इस बार
gommit --set-language hi      # वर्तमान उपयोगकर्ता के लिए सहेजें
```

कोड: `en`, `es`, `pt`, `hi`, `ru`, `zh`। डिफ़ॉल्ट अंग्रेज़ी है।

पहली स्थापना पर Gommit सिस्टम की समर्थित भाषा पहचानकर सहेजता है। `--set-language`
से चुनी गई भाषा कभी बदली नहीं जाती।

### Gemini के साथ स्वचालित कमिट

आधिकारिक Gemini CLI इंस्टॉल और प्रमाणित करें तथा `GEMINI_API_KEY` सेट करें
(`GEMINI_KEY` भी स्वीकार है)। मेनू में `AI` चुनने पर केवल `git diff --cached`
(अधिकतम 512 KiB) से प्रकार और 72 अक्षरों तक तकनीकी विषय बनता है। secrets भेजने
से पहले stage जाँचें; Git/Husky hooks सामान्य रूप से चलते हैं।

Git रेपो (जिसमें बदलाव हों) के अंदर:

```bash
gommit
```

टिपिकल फ़्लो:
1. **type** चुनें (नंबर, नाम, या खोज)
2. **scope** लिखें (वैकल्पिक)
3. **subject** लिखें (आज्ञार्थक, `<=72` अक्षर)
4. **body** जोड़ें (वैकल्पिक, मल्टीलाइन) — अलग लाइन में `.` लिखकर **या** दो बार Enter दबाकर समाप्त करें  
5. **Breaking changes?** (हाँ तो विवरण दें)
6. **Issues?** (Closes/Refs)

अंत में `gommit` `git commit` चलाता है (या `--dry-run` के साथ केवल मैसेज दिखाता है)।

### उदाहरण

सामान्य कमिट (auto‑stage के साथ):
```bash
gommit
```

खाली कमिट की अनुमति:
```bash
gommit --allow-empty
```

पिछला कमिट बदलें:
```bash
gommit --amend
```

सिर्फ मैसेज दिखाएँ (कमिट नहीं):
```bash
gommit --dry-run
```

बिना विज़ार्ड के, फ़्लैग्स से:
```bash
gommit --type feat --scope ui --subject "टाइप सिलेक्टर जोड़ें" --body "इम्प्लीमेंटेशन विवरण..."
```

पूर्ण गैर‑इंटरैक्टिव उदाहरण:
```bash
gommit --type fix --scope api --subject "ऑथेंटिकेशन समस्या सुलझाएँ" --body "JWT टोकन वैलिडेशन ठीक किया

पहले एक्सपायर्ड टोकन सही से हैंडल नहीं हो रहे थे." --footer "Closes #42"
```

---

## 📝 हेडर फ़ॉर्मेट

```
<type>[(!)][(scope)] <emoji>: <subject>
```

उदाहरण:
```
feat 💡: इंस्टॉल कमांड जोड़ें
fix(api) 🐛: कॉन्फ़िग लोड में nil pointer ठीक करें
refactor(core)! 🎨: मैसेज बिल्डर एकरूप करें
```

> `!` केवल **हेडर** में दिखता है; अंदरूनी वैलिडेशन "शुद्ध" टाइप (`feat`, `fix`, आदि) इस्तेमाल करता है ताकि Conventional Commits टूल्स से संगत रहे।

---

## ⚙️ कमांड‑लाइन फ्लैग्स

| फ़्लैग | विवरण |
|---|---|
| `--version` | वर्ज़न दिखाएँ और बाहर निकलें |
| `--install` | यूज़र PATH में बायनरी इंस्टॉल/अपडेट करें और बाहर निकलें |
| `--dry-run` | केवल संदेश प्रिंट करें (`git commit` नहीं) |
| `--language` | इस बार `en`, `es`, `pt`, `hi`, `ru` या `zh` उपयोग करें |
| `--set-language` | उपयोगकर्ता की इंटरफ़ेस भाषा सहेजें |
| `--type` | टाइप (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, WIP, prune) |
| `--scope` | वैकल्पिक scope (जैसे `ui`, `api`) |
| `--subject` | विषय‑पंक्ति (आज्ञार्थक, `<=72` अक्षर) |
| `--body` | बॉडी (नई लाइन के लिए `
`) |
| `--footer` | फूटर (जैसे `Closes #123`) |
| `--as-editor` | Git एडिटर मोड (`COMMIT_EDITMSG`) |
| `--allow-empty` | खाली कमिट की अनुमति |
| `--amend` | पिछला कमिट बदलना |
| `--no-verify` | hooks छोड़ना (`pre-commit`, `commit-msg`) |
| `--signoff` | `Signed-off-by` ट्रेलर जोड़ना |
| `--auto-stage` | जब कुछ भी staged न हो तो `git add -A` (डिफ़ॉल्ट: `true`) |
| `--show-status` | कमिट से पहले staged सारांश दिखाना (डिफ़ॉल्ट: `true`) |

---

## 📋 समर्थित कमिट प्रकार

| टाइप | इमोजी | विवरण |
|---|---|---|
| `WIP` | 🚧 | प्रगति पर |
| `feat` | 💡 | नई सुविधा |
| `fix` | 🐛 | बग ठीक |
| `chore` | 📦 | डिपेंडेंसी/डिप्लॉय/कॉन्फ़िग अपडेट |
| `refactor` | 🎨 | संरचना/फ़ॉर्मेट सुधार (बर्ताव नहीं बदलता) |
| `prune` | 🔥 | कोड/फ़ाइलें हटाना |
| `docs` | 📝 | डाक्यूमेंटेशन |
| `perf` | ⚡ | परफ़ॉर्मेंस सुधार |
| `test` | ✅ | टेस्ट जोड़ना |
| `build` | 🔧 | बिल्ड सिस्टम/डिपेंडेंसी बदलाव |
| `ci` | 🤖 | CI/CD कॉन्फ़िग बदलाव |
| `style` | 💅 | स्टाइल परिवर्तन (अर्थ नहीं बदलता) |
| `revert` | ⏪ | किसी कमिट पर वापस जाना |

---

## 🔧 Git एडिटर मोड

`gommit` को Git एडिटर के रूप में सेट करें:

```bash
git config --global core.editor "gommit --as-editor"
```

जब Git को कमिट मैसेज चाहिए होगा (जैसे `git commit`, `git merge`, `git rebase`), तब `gommit` विज़ार्ड खुलेगा।

---

## 🛠️ डेवलपमेंट

### सोर्स से बिल्ड

आवश्यकता: Go 1.22+

```bash
git clone https://github.com/Hangell/gommit.git
cd gommit

go build -o gommit ./cmd/gommit
go test ./...

./gommit --install
```

### वर्ज़न के साथ बिल्ड

```bash
go build -ldflags "-s -w -X main.version=0.2.0" -o gommit ./cmd/gommit
```

---

## 🎨 एन्वायरनमेंट वेरिएबल्स

| वेरिएबल | विवरण |
|---|---|
| `NO_COLOR=1` | मेन्यू में ANSI रंग बंद करें |
| `NO_EMOJI=1` | मेन्यू में ASCII फॉलबैक (हेडर में फिर भी यूनिकोड इमोजी) |

---

## 🤝 योगदान

1. रेपो Fork करें  
2. ब्रांच बनाएं (`git checkout -b feature/amazing-feature`)  
3. बदलाव `gommit` से कमिट करें 😉  
4. ब्रांच पुश करें (`git push origin feature/amazing-feature`)  
5. Pull Request खोलें

---

## 📄 लाइसेंस

GPL‑3.0‑only — विवरण के लिए [`LICENSE`](LICENSE) देखें।

---

## ❓ FAQ

**क्या gommit, Commitizen का विकल्प है?** यह **विकल्प** है। यदि आपका सेटअप Node/npm और cz एडाप्टर पर टिका है तो आप cz जारी रख सकते हैं। यदि आप **कम निर्भरताएँ**, **एक बार इंस्टॉलेशन** और कई रेपो/सर्वर पर **समान अनुभव** चाहते हैं, तो gommit अधिक सरल है।

**क्या यह तेज़ है?** यहाँ बेंचमार्क नहीं, पर नैटिव बायनरी होने से gommit आमतौर पर **तेज़ स्टार्ट** और कम मेमोरी देता है—विशेषकर CI/कंटेनर के “कोल्ड स्टार्ट” में, जहाँ Node स्टार्ट‑अप देरी जोड़ता है।

**ऑफ़लाइन चलता है?** हाँ। इंस्टॉल हो जाने के बाद नेटवर्क की ज़रूरत नहीं।

**gommit क्यों बना?** nvm/प्रोजेक्ट‑दर‑प्रोजेक्ट cz इंस्टॉल की झंझट से बचने के लिए—एक सरल, एकीकृत टूल प्रदान करने हेतु।

---

## 👨‍💻 लेखक

<div align="center">
  <img src="https://avatars.githubusercontent.com/u/53544561?v=4" width="150" style="border-radius: 50%;" />

**Rodrigo Rangel**

  <div>
    <a href="https://hangell.org" target="_blank">
      <img src="https://img.shields.io/badge/website-000000?style=for-the-badge&logo=About.me&logoColor=white" alt="Website" />
    </a>
    <a href="https://play.google.com/store/apps/dev?id=5606456325281613718" target="_blank">
      <img src="https://img.shields.io/badge/Google_Play-414141?style=for-the-badge&logo=google-play&logoColor=white" alt="Google Play" />
    </a>
    <a href="https://www.youtube.com/channel/UC8_zG7RFM2aMhI-p-6zmixw" target="_blank">
      <img src="https://img.shields.io/badge/YouTube-FF0000?style=for-the-badge&logo=youtube&logoColor=white" alt="YouTube" />
    </a>
    <a href="https://www.facebook.com/hangell.org" target="_blank">
      <img src="https://img.shields.io/badge/Facebook-1877F2?style=for-the-badge&logo=facebook&logoColor=white" alt="Facebook" />
    </a>
    <a href="https://www.linkedin.com/in/rodrigo-rangel-a80810170" target="_blank">
      <img src="https://img.shields.io/badge/-LinkedIn-%230077B5?style=for-the-badge&logo=linkedin&logoColor=white" alt="LinkedIn" />
    </a>
  </div>
</div>

---

## 🙏 आभार

- [Commitizen](https://github.com/commitizen/cz-cli) से प्रेरित
- [Conventional Commits](https://www.conventionalcommits.org/) स्पेसिफिकेशन पर आधारित
- इमोजी कमिट इतिहास को और दृश्यात्मक व मज़ेदार बनाते हैं! 🎉

---

<div align="center">
  <strong>डेवलपर समुदाय के लिए ❤️ के साथ</strong>
</div>
