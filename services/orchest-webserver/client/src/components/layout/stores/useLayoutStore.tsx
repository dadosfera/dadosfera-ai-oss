import { clamp } from "@/utils/math";
import { create } from "zustand";
import { persist } from "zustand/middleware";

type LayoutStateAction = number | ((prevValue: number) => number);

const getValue = (currentValue: number, value: LayoutStateAction) =>
  value instanceof Function ? value(currentValue) : value;

export type LayoutState = {
  mainSidePanelWidth: number;
  setMainSidePanelWidth: (value: LayoutStateAction) => void;
  secondarySidePanelWidth: number;
  setSecondarySidePanelWidth: (value: LayoutStateAction) => void;
};

const DEFAULT_MAIN_SIDE_PANEL_WIDTH = 300;
export const MIN_MAIN_SIDE_PANEL_WIDTH = 280;
const DEFAULT_SECONDARY_SIDE_PANEL_WIDTH = 450;
export const MIN_SECONDARY_SIDE_PANEL_WIDTH = 420;

export const useLayoutStore = create(
  persist<LayoutState>(
    (set) => ({
      mainSidePanelWidth: DEFAULT_MAIN_SIDE_PANEL_WIDTH,
      setMainSidePanelWidth: (value) => {
        set(({ mainSidePanelWidth }) => ({
          mainSidePanelWidth: clamp(
            getValue(mainSidePanelWidth, value),
            MIN_MAIN_SIDE_PANEL_WIDTH,
            window.innerWidth / 2
          ),
        }));
      },
      secondarySidePanelWidth: DEFAULT_SECONDARY_SIDE_PANEL_WIDTH,
      setSecondarySidePanelWidth: (value) => {
        set(({ secondarySidePanelWidth }) => ({
          secondarySidePanelWidth: clamp(
            getValue(secondarySidePanelWidth, value),
            MIN_SECONDARY_SIDE_PANEL_WIDTH,
            window.innerWidth / 2
          ),
        }));
      },
    }),
    { name: "orchest.layout" }
  )
);

export const useMainSidePanelWidth = () => {
  const mainSidePanelWidth = useLayoutStore(
    (state) => state.mainSidePanelWidth
  );
  const setMainSidePanelWidth = useLayoutStore(
    (state) => state.setMainSidePanelWidth
  );

  return [mainSidePanelWidth, setMainSidePanelWidth] as const;
};

export const useSecondarySidePanelWidth = () => {
  const secondarySidePanelWidth = useLayoutStore(
    (state) => state.secondarySidePanelWidth
  );
  const setSecondarySidePanelWidth = useLayoutStore(
    (state) => state.setSecondarySidePanelWidth
  );
  return [secondarySidePanelWidth, setSecondarySidePanelWidth] as const;
};
