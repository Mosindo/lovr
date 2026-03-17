import React from "react";
import { Card as TamaguiCard, type GetProps } from "tamagui";

export type CardProps = GetProps<typeof TamaguiCard>;

export default function Card(props: CardProps) {
  return <TamaguiCard {...props} />;
}
