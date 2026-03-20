import React from "react";
import { Input as TamaguiInput, type GetProps } from "tamagui";

export type InputProps = GetProps<typeof TamaguiInput>;

export default function Input(props: InputProps) {
  return <TamaguiInput {...props} />;
}
