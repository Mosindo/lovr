import React from "react";
import { Button as TamaguiButton, type GetProps } from "tamagui";

export type ButtonProps = GetProps<typeof TamaguiButton>;

export default function Button(props: ButtonProps) {
  return <TamaguiButton {...props} />;
}
