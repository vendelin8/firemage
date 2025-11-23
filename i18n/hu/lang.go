package internal

const (
	menuRefresh    = "Frissít"
	menuSave       = "Ment"
	menuCancel     = "Mégse"
	menuQuit       = "Kilép"
	shortDesc      = "Firebase jogosultságkezelő"
	longDesc       = "firemage egy Go nyelven készült Firebase jogosultságokat kezelő terminál alkalmazás"
	descKey        = "Google service account fájl útvonala"
	descConf       = "beállítás fájl útvonala"
	descDebug      = "hibakereső üzenetek"
	descEmul       = "helyi firebase emulátor használata"
	sName          = "Név"
	sEmail         = "Email"
	sSearchThis    = "Keresés erre:"
	sDoSearch      = "Keresés"
	sYes           = "Igen"
	sNo            = "Nem"
	sSaved         = "Mentés OK"
	sNoChanges     = "Nem történt változás"
	cShortcuts     = "gyorsbillentyűk"
	warnUnsaved    = "%d darab nem mentett akciód van. Biztos kilépsz?"
	warnNoUsers    = "Nincs idevágó felhasználó"
	errCantRefresh = "A frissítés csak a Lista oldalon lehetséges!"
	errActions     = "Először mentsd el az aktuális nézetet, vagy vond vissza a Mégse gombbal!"
	errMinLen      = "Legalább %d karaktert írj be!"
	warnMayRefresh = "Fontold meg a frissítést!"
	errManual      = "Valaki kézzel belenyúlt a jogokba vagy az adatbázisba?"
	errNew         = "Új felhasználó(k) a jogosultak között: %s ."
	errChanged     = "Megváltozott jogosultságú felhasználó(k): %s ."
	errEmpty       = "Neki(k) megszűntek a jogosultságai(k): %s ."
	errTimeout     = "Adatabázis időkorlát túllépés."
	errRemoved     = "Az alábbiak törlődtek a rendszerből: %s ."
	errConfPath    = "Nincs meg a beállítás fájl, ellenőrizd a program paramétereit."
	errConfParse   = "Beállítás fájl hibás: %s ."
	errCmdNotFound = "Hiányzó gyorsbillentyű parancs(ok): %s ."
	errKeyNotFound = "Hiányzó gyorsbillentyű(k): %s ."
)

var (
	titles = map[string]string{pageSrch: "Kereső", pageLst: "Lista"}
	warns  = map[int]string{
		wSearchAgain:  "A változtatásaid megmaradnak az előző keresésből. Ha mégse szeretnéd őket, nyomj a Mégse gombra.",
		wActionInList: "A korábbi változásaid megmaradnak. Ha a keresésnél hozzáadtál valakit, itt csak mentés után fogod látni.",
	}
)
