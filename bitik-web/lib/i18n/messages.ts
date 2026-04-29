export type Locale = "en" | "ar"

export const localeDirection: Record<Locale, "ltr" | "rtl"> = {
  en: "ltr",
  ar: "rtl",
}

export const messages: Record<Locale, Record<string, string>> = {
  en: {
    "common.language": "Language",
    "nav.categories": "Categories",
    "nav.brands": "Brands",
    "nav.products": "Products",
    "nav.signIn": "Sign in",
    "nav.searchLabel": "Search Bitik",
    "nav.searchPlaceholder": "Search products, brands, sellers",
    "offline.message": "You are offline. Some actions may be unavailable.",
    "session.stale": "Session expired. Please sign in again.",
  },
  ar: {
    "common.language": "اللغة",
    "nav.categories": "الفئات",
    "nav.brands": "العلامات",
    "nav.products": "المنتجات",
    "nav.signIn": "تسجيل الدخول",
    "nav.searchLabel": "ابحث في بيتيك",
    "nav.searchPlaceholder": "ابحث عن المنتجات والعلامات والبائعين",
    "offline.message": "أنت غير متصل. قد لا تعمل بعض الإجراءات.",
    "session.stale": "انتهت الجلسة. يرجى تسجيل الدخول مرة أخرى.",
  },
}

