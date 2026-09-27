package entity

// MenuBrowse is the Phase 1 HTTP read model for one source store menu path.
type MenuBrowse struct {
	Store      Store
	Menu       Menu
	Categories []CategoryWithItems
}

type CategoryWithItems struct {
	Category Category
	Items    []MenuItemView
}

type MenuItemView struct {
	Item       Item
	PriceCents int64
	Currency   string
}
