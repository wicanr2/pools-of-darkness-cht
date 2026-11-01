package pod

import "bytes"

// HDLayers 是逐幀的牆面、戰鬥與遭遇圖呈現層。主專案規格 010 §6、§7。
// 它只讀原版色號，呼叫端在既有文字與面板之前繪製。
type HDLayers struct {
	combat      *combatHD
	actors      *actorHD
	encounters  *encounterHD
	walls       *wallHD
	exact       *exactHD
	composite   *combatComposite
	classified  bool
	lastIndex   []byte
	lastPalette [256][3]uint8
}

func LoadHDLayers(dir string) (*HDLayers, int, error) {
	combat, err := loadCombatHD(dir)
	if err != nil {
		return nil, 0, err
	}
	actors, err := loadActorHD(dir)
	if err != nil {
		return nil, 0, err
	}
	characters, err := loadCharacterHD(dir)
	if err != nil {
		return nil, 0, err
	}
	actors = mergeCharacterHD(actors, characters)
	effects, err := loadEffectHD(dir)
	if err != nil {
		return nil, 0, err
	}
	actors, err = mergeEffectHD(actors, effects)
	if err != nil {
		return nil, 0, err
	}
	encounters, err := loadEncounterHD(dir)
	if err != nil {
		return nil, 0, err
	}
	walls, err := loadWallHD(dir)
	if err != nil {
		return nil, 0, err
	}
	exact, err := loadExactHD(dir)
	if err != nil {
		return nil, 0, err
	}
	count := 0
	if combat != nil {
		count += len(combat.tiles)
	}
	if actors != nil {
		count += actors.sources
	}
	if encounters != nil {
		count += len(encounters.items)
	}
	if walls != nil {
		count += walls.sources
	}
	if exact != nil {
		count += len(exact.items)
	}
	return &HDLayers{combat: combat, actors: actors, encounters: encounters, walls: walls, exact: exact, composite: newCombatComposite(combat, actors)}, count, nil
}

func (h *HDLayers) Frame(indexed []byte, pal [256][3]uint8) {
	if h == nil {
		return
	}
	if len(indexed) != 320*200 {
		h.classified = false
		h.classify(indexed, pal)
		return
	}
	if h.classified && pal == h.lastPalette && bytes.Equal(indexed, h.lastIndex) {
		return
	}
	h.classify(indexed, pal)
	h.lastIndex = append(h.lastIndex[:0], indexed...)
	h.lastPalette, h.classified = pal, true
}

// classify depends only on the immutable theme, source indices and palette.
// Presentation holds advance in Live.Frame, outside this classification cache.
func (h *HDLayers) classify(indexed []byte, pal [256][3]uint8) {
	h.combat.frame(indexed, pal)
	guard := h.combat != nil && len(h.combat.matches) >= 3
	if h.actors != nil {
		h.actors.vgaPaletteOK = paletteBankHash(pal, 144, 256) == cpicVGAPalette
		h.actors.effectPaletteOK = paletteBankHash(pal, 64, 96) == comsprVGAPalette
	}
	h.actors.frame(indexed, guard)
	h.actors.partialFrame(indexed, guard)
	h.composite.frame(indexed, pal, h.combat, h.actors)
	h.encounters.frame(indexed, pal)
	h.walls.frame(indexed, pal)
	h.exact.frame(indexed, pal)
}

func (h *HDLayers) pendingBattleCells() (cells [21 * 21]bool) {
	if h == nil || h.combat == nil || h.combat.vgaGuards < 3 {
		return
	}
	if len(h.combat.matches) < 3 || h.actors != nil && !h.actors.vgaPaletteOK {
		for n := range cells {
			cells[n] = true
		}
	} else if h.actors != nil {
		cells = h.actors.deferCells
	}
	if h.composite != nil {
		for n, pending := range h.composite.deferCells {
			cells[n] = cells[n] || pending
		}
	}
	return
}

func (h *HDLayers) Draw(out, indexed []byte, scale int) {
	if h == nil {
		return
	}
	h.walls.draw(out, indexed, scale)
	h.combat.draw(out, indexed, scale)
	h.composite.drawTerrain(out, indexed, scale, h.combat)
	h.actors.draw(out, indexed, scale)
	h.composite.drawActors(out, indexed, scale, h.actors)
	h.encounters.draw(out, indexed, scale)
	h.exact.draw(out, indexed, scale)
}
