export interface Vote {
    model_slug: string;
    present: boolean;
    evidence: string;
    explanation: string;
}

export interface Cell {
    individual_id: number;
    pattern_id: number;
    code: string;
    name: string;
    review_text: string;
    panel_vote: boolean;
    n_present: number;
    n_total: number;
    votes: Vote[];
}

export interface WorklistItem {
    individual_id: number;
    pattern_id: number;
    code: string;
    decided: boolean;
}

export interface DecisionInput {
    gold_run: number;
    panel_run: number;
    individual_id: number;
    pattern_id: number;
    label: boolean;
}

export interface Detection {
    model_slug: string;
    evidence: string;
    explanation: string;
}

export interface ReviewPattern {
    pattern_id: number;
    code: string;
    name: string;
    family: string;
    description: string;
    panel_vote: boolean;
    n_present: number;
    n_total: number;
    decided: boolean;
    final_label: boolean;
    detections: Detection[];
}

export interface Review {
    individual_id: number;
    review_text: string;
    n_total: number;
    patterns: ReviewPattern[];
}

export interface ReviewItem {
    individual_id: number;
    external_game_id: number;
    decided: number;
    total: number;
}

export interface ReviewDecisionInput {
    gold_run: number;
    panel_run: number;
    decisions: { pattern_id: number; label: boolean }[];
}