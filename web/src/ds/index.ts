/* GadgetCast design system, ported to typed TSX from the system's components/core + components/studio.
   Pixels only: no data fetching, no app state. Colours, type, spacing and motion resolve through tokens/. */
export { Icon, type IconProps, type IconSize } from "./core/Icon";
export { Button, type ButtonProps, type ButtonVariant } from "./core/Button";
export { IconButton, type IconButtonProps } from "./core/IconButton";
export { Badge, type BadgeProps, type BadgeTone } from "./core/Badge";
export { Card, type CardProps } from "./core/Card";
export { Input, type InputProps } from "./core/Input";
export { Select, type SelectProps, type SelectOption } from "./core/Select";
export { Switch, type SwitchProps } from "./core/Switch";
export { Slider, type SliderProps } from "./core/Slider";
export { Tabs, type TabsProps, type TabItem } from "./core/Tabs";
export { Modal, type ModalProps } from "./core/Modal";
export { Tooltip, type TooltipProps } from "./core/Tooltip";
export { Menu, type MenuProps, type MenuItem } from "./core/Menu";
export { Toast, type ToastProps, type ToastTone } from "./core/Toast";
export { PortHandle, SIGNAL_HUE, type PortHandleProps, type SignalKind } from "./studio/PortHandle";
export { PipelineNode, NODE_CLASS_HUE, type PipelineNodeProps, type NodePort, type NodeClass } from "./studio/PipelineNode";
export { PatchWire, wirePath, type PatchWireProps } from "./studio/PatchWire";
export { LevelMeter, WaveScope, OnAirIndicator, MetricPill } from "./studio/Meters";
export { LogStream, PresetRow, DeviceRow, Table, type LogLine, type LogLevel, type TableColumn, type DeviceStatus } from "./studio/Lists";
